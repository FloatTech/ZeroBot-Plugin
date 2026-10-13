//go:build ignore

// Run from the repository root: go run scripts/smoke-local-admin.go
// Uses synthetic accounts and a fresh fixture directory, never the live config.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode/utf8"

	"github.com/RomiChan/websocket"
)

type apiRequest struct {
	Action string `json:"action"`
	Params struct {
		UserID    int64           `json:"user_id"`
		GroupID   int64           `json:"group_id"`
		Message   json.RawMessage `json:"message"`
		Messages  json.RawMessage `json:"messages"`
		Character string          `json:"character"`
		Text      string          `json:"text"`
	} `json:"params"`
	Echo json.RawMessage `json:"echo"`
}

func main() {
	flag.Parse()
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var uiWait = flag.Bool("ui-wait", false, "Wait for local browser QA after integration checks.")
var binaryName = flag.String("binary-name", "zerobot.exe", "Windows binary filename under .local/bin.")

func run() error {
	root, err := os.Getwd()
	if err != nil {
		return err
	}
	if filepath.Base(*binaryName) != *binaryName || filepath.Ext(*binaryName) != ".exe" {
		return fmt.Errorf("binary-name must be an exe filename")
	}
	artifactRoot := filepath.Join(root, ".local", "smoke-admin")
	if err := os.MkdirAll(artifactRoot, 0700); err != nil {
		return err
	}
	runtimeRoot, err := os.MkdirTemp(artifactRoot, "run-")
	if err != nil {
		return err
	}
	if err := prepareFixture(root, runtimeRoot); err != nil {
		return err
	}
	var modelMu sync.Mutex
	modelRequests := []map[string]any{}
	shortModelReply := "本地模型模拟回复🙂" + strings.Repeat("中", 200-utf8.RuneCountInString("本地模型模拟回复🙂"))
	longModelReply := shortModelReply + "😀"
	modelServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer synthetic-model-key" {
			http.Error(w, "fake key mismatch", 401)
			return
		}
		var body map[string]any
		if json.NewDecoder(r.Body).Decode(&body) != nil {
			http.Error(w, "bad model request", 400)
			return
		}
		modelMu.Lock()
		modelRequests = append(modelRequests, body)
		modelMu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		bodyJSON, _ := json.Marshal(body)
		reply := shortModelReply
		if body["model"] == "smoke-voice-chat" {
			reply = "VOICE_TEXT_OK"
		}
		if bytes.Contains(bodyJSON, []byte("quoted smoke")) {
			reply = longModelReply
		}
		if body["model"] == "smoke-role-agent" {
			kind := "user"
			messages, _ := body["messages"].([]any)
			if len(messages) > 0 {
				system, _ := messages[0].(map[string]any)
				prompt, _ := system["content"].(string)
				if strings.Contains(prompt, "|群组踢人|set_group_kick|") {
					kind = "admin"
				}
				if strings.Contains(prompt, "|群组设置管理员|set_group_admin|") {
					kind = "owner"
				}
			}
			text := "ROLE_AGENT_" + kind
			pre := ""
			if strings.LastIndex(string(bodyJSON), "ROLE_AGENT_CROSS_INFO") > strings.LastIndex(string(bodyJSON), "ROLE_AGENT_CHECK") {
				pre = `{"action":"get_group_info","params":{"group_id":202}}` + "\n"
				text = "ROLE_BLOCKED_OK"
			}
			if strings.LastIndex(string(bodyJSON), "ROLE_AGENT_CQ_ALL") > max(strings.LastIndex(string(bodyJSON), "ROLE_AGENT_CHECK"), strings.LastIndex(string(bodyJSON), "ROLE_AGENT_CROSS_INFO")) {
				pre = `{"action":"send_group_msg","params":{"group_id":303,"message":"[CQ:at,qq=all]UNAUTHORIZED"}}` + "\n"
				text = "ROLE_BLOCKED_OK"
			}
			reply = pre + `{"action":"send_group_msg","params":{"group_id":303,"message":"` + text + `"}}` + "\n" + `{"action":"end_action","params":{}}`
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"choices": []any{map[string]any{"message": map[string]any{"role": "assistant", "content": reply}}}})
	}))
	defer modelServer.Close()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	defer listener.Close()
	connected := make(chan *websocket.Conn, 1)
	sent := make(chan apiRequest, 8)
	var failServiceForward atomic.Bool
	var roleCrossGroupQuery atomic.Bool
	var voiceGets, voiceDirects, voiceRecordAttempts atomic.Int32
	var voiceEmptyRecord, voiceInvalidTarget atomic.Bool
	var writeMu sync.Mutex
	server := &http.Server{ReadHeaderTimeout: 5 * time.Second}
	upgrader := websocket.Upgrader{}
	server.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer local-smoke-only" {
			http.Error(w, "unauthorized", 401)
			return
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		writeMu.Lock()
		err = conn.WriteJSON(map[string]any{"self_id": 87654321, "post_type": "meta_event", "meta_event_type": "lifecycle", "sub_type": "connect", "time": time.Now().Unix()})
		writeMu.Unlock()
		if err != nil {
			return
		}
		connected <- conn
		for {
			var request apiRequest
			if err := conn.ReadJSON(&request); err != nil {
				return
			}
			data := map[string]any{"message_id": 999, "user_id": 12345678, "nickname": "Local smoke user"}
			var apiData any = data
			if request.Action == "get_ai_characters" {
				apiData = []any{map[string]any{"characters": []any{map[string]string{"character_name": "Synthetic voice", "character_id": "synthetic-voice-1"}, map[string]string{"character_name": "Synthetic voice", "character_id": "synthetic-voice-2"}}}, map[string]any{"characters": []any{map[string]string{"character_name": "Repeated voice", "character_id": "synthetic-voice-1"}}}}
			}
			voiceAccepted := false
			apiStatus, apiRetcode := "ok", 0
			if request.Action == "get_ai_record" {
				// Installed NapCat posts to the group before returning the URL.
				voiceGets.Add(1)
				apiData = "https://synthetic.voice/record"
				voiceAccepted = true
			}
			if request.Action == "send_group_ai_record" {
				voiceDirects.Add(1)
				if (request.Params.GroupID != 303 && request.Params.GroupID != 404) || request.Params.Character != "synthetic-voice-1" {
					voiceInvalidTarget.Store(true)
				}
				// Native successful delivery has no message ID.
				data["message_id"] = 0
				voiceAccepted = true
				if request.Params.Text == "group-failure" {
					apiStatus, apiRetcode, voiceAccepted = "failed", 1400, false
				}
			}
			failedVoiceSend := false
			if request.Action == "send_private_msg" || request.Action == "send_group_msg" {
				var segments []struct {
					Type string            `json:"type"`
					Data map[string]string `json:"data"`
				}
				_ = json.Unmarshal(request.Params.Message, &segments)
				for _, segment := range segments {
					if segment.Type == "record" {
						voiceRecordAttempts.Add(1)
						if segment.Data["file"] == "" {
							voiceEmptyRecord.Store(true)
						}
						if segment.Data["file"] == "https://synthetic.voice/fail" {
							failedVoiceSend = true
							data["message_id"] = 0
						}
					}
				}
			}
			if request.Action == "get_group_info" && request.Params.GroupID == 202 {
				roleCrossGroupQuery.Store(true)
			}
			failedForward := strings.HasSuffix(request.Action, "forward_msg") && failServiceForward.CompareAndSwap(true, false)
			if failedForward {
				data["message_id"] = 0
			}
			if request.Action == "get_msg" {
				data["message"] = []map[string]any{{"type": "text", "data": map[string]string{"text": "quoted context"}}}
				data["message_type"] = "private"
				data["sender"] = map[string]any{"user_id": 12345678, "nickname": "Quoted smoke user"}
			}
			writeMu.Lock()
			err := conn.WriteJSON(map[string]any{"status": apiStatus, "retcode": apiRetcode, "data": apiData, "echo": request.Echo})
			writeMu.Unlock()
			if err != nil {
				return
			}
			if voiceAccepted {
				native := request
				native.Action = "send_group_ai_record"
				native.Params.Message = json.RawMessage(`[{"type":"record","data":{"file":"synthetic-native-record"}}]`)
				sent <- native
			}
			if !failedForward && !failedVoiceSend && (request.Action == "send_private_msg" || request.Action == "send_group_msg" || request.Action == "send_private_forward_msg" || request.Action == "send_group_forward_msg") {
				sent <- request
			}
		}
	})
	go server.Serve(listener)
	defer server.Close()
	config := map[string]any{
		"zero": map[string]any{"nickname": []string{"七叶叶"}, "command_prefix": "/", "super_users": []int64{12345678}, "ring_len": 4096, "latency": 233000000, "max_process_time": 240000000000, "mark_message": true},
		"ws":   []map[string]any{{"Url": "ws://" + listener.Addr().String(), "AccessToken": "local-smoke-only"}}, "wss": nil, "force_base64_file": false,
	}
	encoded, err := json.Marshal(config)
	if err != nil {
		return err
	}
	configPath := filepath.Join(runtimeRoot, "config.json")
	if err := os.WriteFile(configPath, encoded, 0600); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	adminListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	adminAddress := adminListener.Addr().String()
	adminListener.Close()
	adminData := filepath.Join(runtimeRoot, "admin-"+time.Now().Format("20060102-150405"))
	cmd := exec.CommandContext(ctx, filepath.Join(root, ".local", "bin", *binaryName), "-c", configPath, "-admin", "-admin-addr", adminAddress, "-admin-data", adminData)
	cmd.Dir = runtimeRoot
	cmd.Env = append(os.Environ(), "DEBUG_MODE=1")
	stdout, err := os.Create(filepath.Join(runtimeRoot, "stdout.log"))
	if err != nil {
		return err
	}
	defer stdout.Close()
	stderr, err := os.Create(filepath.Join(runtimeRoot, "stderr.log"))
	if err != nil {
		return err
	}
	defer stderr.Close()
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if err := cmd.Start(); err != nil {
		return err
	}
	exited := make(chan error, 1)
	go func() { exited <- cmd.Wait() }()
	defer func() {
		cancel()
		select {
		case <-exited:
		case <-time.After(3 * time.Second):
		}
	}()
	var conn *websocket.Conn
	select {
	case conn = <-connected:
	case err := <-exited:
		return fmt.Errorf("isolated bot exited before connecting: %v", err)
	case <-ctx.Done():
		return fmt.Errorf("isolated bot startup timed out")
	}
	defer conn.Close()
	// Let RunAndBlock finish installing the event listener after the handshake.
	time.Sleep(500 * time.Millisecond)
	controlMessageID := 100
	sendMessage := func(uid, gid int64, role, command string, atBot bool) (string, error) {
		controlMessageID++
		segments := []map[string]any{{"type": "text", "data": map[string]string{"text": command}}}
		// This one fixture command exercises QQ reply-element preprocessing.
		if command == "/gpt quoted smoke" {
			segments = append([]map[string]any{{"type": "reply", "data": map[string]string{"id": "42"}}}, segments...)
		}
		if atBot {
			segments = append([]map[string]any{{"type": "at", "data": map[string]string{"qq": "87654321"}}}, segments...)
		}
		event := map[string]any{"time": time.Now().Unix(), "self_id": 87654321, "post_type": "message", "message_type": "private", "sub_type": "friend", "message_id": controlMessageID, "user_id": uid, "raw_message": command, "message": segments, "sender": map[string]any{"user_id": uid, "nickname": "Synthetic permission user", "sex": "unknown", "age": 0, "role": role}}
		if gid != 0 {
			event["message_type"], event["sub_type"], event["group_id"] = "group", "normal", gid
		}
		writeMu.Lock()
		err := conn.WriteJSON(event)
		writeMu.Unlock()
		if err != nil {
			return "", err
		}
		select {
		case reply := <-sent:
			if (gid == 0 && (reply.Params.UserID != uid || !strings.HasPrefix(reply.Action, "send_private_"))) || (gid != 0 && (reply.Params.GroupID != gid || !strings.HasPrefix(reply.Action, "send_group_"))) {
				return "", fmt.Errorf("unexpected command recipient")
			}
			raw := reply.Params.Message
			if (gid > 0 && strings.HasPrefix(command, "发送AI语音") && command != "发送AI语音group-failure") || command == "VOICE_CHAT_GROUP" {
				var segments []struct {
					Type string            `json:"type"`
					Data map[string]string `json:"data"`
				}
				if json.Unmarshal(raw, &segments) != nil || len(segments) != 1 || segments[0].Type != "record" || segments[0].Data["file"] == "" {
					return "", fmt.Errorf("voice command did not send one nonempty record")
				}
				select {
				case <-sent:
					return "", fmt.Errorf("native voice command produced an extra reply")
				case <-time.After(200 * time.Millisecond):
				}
				return "VOICE_OK", nil
			}
			if strings.HasSuffix(reply.Action, "forward_msg") {
				raw = reply.Params.Messages
			}
			if command == "/gpt group smoke" || command == "/gpt quoted smoke" {
				expectedAction, expectedText := "send_group_msg", shortModelReply
				if command == "/gpt quoted smoke" {
					expectedAction, expectedText = "send_private_forward_msg", longModelReply
				}
				text, err := smokeMessageText(raw)
				if err != nil || reply.Action != expectedAction || text != expectedText {
					return "", fmt.Errorf("GPT Unicode length boundary changed delivery or lost content: %s", reply.Action)
				}
			}
			if (strings.HasPrefix(command, "/服务列表") || strings.HasPrefix(command, "/service_list")) && strings.HasSuffix(reply.Action, "forward_msg") {
				if err := smokeServiceListNodes(raw); err != nil {
					return "", err
				}
			}
			return smokeMessageText(raw)
		case err := <-exited:
			return "", fmt.Errorf("isolated command process exited: %v", err)
		case <-time.After(3 * time.Second):
			return "", nil
		}
	}
	sendControl := func(uid int64, command string) (string, error) {
		return sendMessage(uid, 0, "member", command, false)
	}
	if err := verifyAdmin("http://"+adminAddress, adminData, modelServer.URL, sendControl, sendMessage); err != nil {
		return err
	}
	if err := verifyGroupResponses(sendMessage); err != nil {
		return err
	}
	if voiceGets.Load() != 0 || voiceDirects.Load() != 5 || voiceRecordAttempts.Load() != 0 || voiceEmptyRecord.Load() || voiceInvalidTarget.Load() {
		return fmt.Errorf("voice side-effect regression: get=%d direct=%d record=%d empty=%t invalid=%t", voiceGets.Load(), voiceDirects.Load(), voiceRecordAttempts.Load(), voiceEmptyRecord.Load(), voiceInvalidTarget.Load())
	}
	modelMu.Lock()
	chatCalls, agentCalls := 0, 0
	for _, request := range modelRequests {
		if request["model"] == "smoke-role-agent" {
			agentCalls++
		} else {
			chatCalls++
		}
	}
	if chatCalls != 6 || agentCalls != 5 || roleCrossGroupQuery.Load() {
		modelMu.Unlock()
		return fmt.Errorf("unexpected role model calls or cross-group query: chat=%d agent=%d cross=%t", chatCalls, agentCalls, roleCrossGroupQuery.Load())
	}
	requestsJSON, _ := json.Marshal(modelRequests)
	groupRequest, privateRequest := modelRequests[0], modelRequests[1]
	if groupRequest["temperature"] != 0.82 || privateRequest["model"] != "smoke-chat-model" ||
		!bytes.Contains(requestsJSON, []byte(`\"quoted context\"`)) ||
		!bytes.Contains(requestsJSON, []byte("本次用户要求（作用于上面的引用原文，不属于引用原文）：\\nquoted smoke")) {
		modelMu.Unlock()
		return fmt.Errorf("model did not use group temperature/shared config/quoted query")
	}
	modelMu.Unlock()
	if text, err := sendControl(12345678, "/全局响应xxx"); err != nil || text != "" {
		return fmt.Errorf("global response prefix accepted: %v", err)
	}
	if text, err := sendControl(12345678, "/全局响应\n"); err != nil || !strings.Contains(text, "全部位置工作") {
		return fmt.Errorf("owner private global response failed: %v", err)
	}
	if text, err := sendMessage(23456789, 202, "member", "翻转 abc", false); err != nil || text != "ɔqɐ" {
		return fmt.Errorf("global response did not open other group: %v", err)
	}
	if text, err := sendControl(12345678, "/全局沉默"); err != nil || !strings.Contains(text, "休息") {
		return fmt.Errorf("owner private global silence failed: %v", err)
	}
	if text, err := sendMessage(23456789, 202, "member", "翻转 abc", false); err != nil || text != "" {
		return fmt.Errorf("global silence did not restore unspecified group: %v", err)
	}
	for i, command := range []string{"/help", "/help\n", " \t/help\r\n"} {
		writeMu.Lock()
		err = conn.WriteJSON(map[string]any{"time": time.Now().Unix(), "self_id": 87654321, "post_type": "message", "message_type": "private", "sub_type": "friend", "message_id": i + 1, "user_id": 12345678, "raw_message": command, "message": []map[string]any{{"type": "text", "data": map[string]string{"text": command}}}, "sender": map[string]any{"user_id": 12345678, "nickname": "Local smoke user", "sex": "unknown", "age": 0}})
		writeMu.Unlock()
		if err != nil {
			return err
		}
		var request apiRequest
		select {
		case request = <-sent:
		case err := <-exited:
			return fmt.Errorf("isolated bot exited during %s: %v", command, err)
		case <-time.After(15 * time.Second):
			return fmt.Errorf("no private response to %s", command)
		}
		if request.Params.UserID != 12345678 {
			return fmt.Errorf("wrong response recipient")
		}
		var segments []struct {
			Type string            `json:"type"`
			Data map[string]string `json:"data"`
		}
		if err := json.Unmarshal(request.Params.Message, &segments); err != nil {
			return err
		}
		valid := false
		for _, segment := range segments {
			if strings.TrimSpace(command) == "/help" && segment.Type == "text" && strings.Contains(segment.Data["text"], "OneBot + ZeroBot + Golang") {
				valid = true
			}
		}
		if !valid {
			return fmt.Errorf("unexpected response payload to %s", command)
		}
	}
	failServiceForward.Store(true)
	if text, err := sendControl(12345678, "/服务列表"); err != nil || !strings.Contains(text, "服务列表聊天记录发送失败") || failServiceForward.Load() {
		return fmt.Errorf("service list failed forward did not give truthful feedback: %v", err)
	}
	if err = verifyReverseDriver(root, runtimeRoot, *binaryName); err != nil {
		return err
	}
	result := map[string]any{"passed": true, "isolated": true, "realQQMessagesSent": false, "helpTextResponse": true, "helpWhitespaceResponse": true, "serviceListForwardUsagePassed": true, "adminAuthPassed": true, "adminActualRegistryPassed": true, "adminEnabledCommandRepliesPassed": true, "adminConnectionsSaveBackupPassed": true, "adminPluginSwitchPassed": true, "chatWebSharedStatePassed": true, "pluginMessageGatePassed": true, "nonOwnerGlobalCommandRejected": true, "completedAt": time.Now().Format(time.RFC3339)}
	result["groupResponseMentionRequiredPassed"] = true
	result["groupSilenceMentionRequiredPassed"] = true
	result["groupResponsePermissionsPassed"] = true
	result["responseCommandBoundaryPassed"] = true
	result["responseLocationIsolationPassed"] = true
	result["ownerPrivateGlobalResponsePassed"] = true
	result["aiSharedConfigPassed"] = true
	result["aiConfigChecksPassed"] = true
	result["aiSecretRedactionPassed"] = true
	result["aiGroupParametersPassed"] = true
	result["gptMockReplyAndQuotePassed"] = true
	result["gptQuoteBoundaryPassed"] = true
	result["aiRoleCommandsPassed"] = true
	result["aiAgentRolePromptsPassed"] = true
	result["aiAgentGroupIsolationPassed"] = true
	result["aiAgentMemberAtAllDeniedPassed"] = true
	result["qqVoiceDeliveryPassed"] = true
	result["qqVoiceNativeDirectPassed"] = true
	result["qqVoicePrivateNoGroupSendPassed"] = true
	result["qqVoiceChatOutputScopePassed"] = true
	result["qqVoiceZeroIDReceiptPassed"] = true
	result["qqVoiceNoResourceForwardingPassed"] = true
	result["qqVoiceFailureAndScopePassed"] = true
	result["qqVoiceModelSelectionPassed"] = true
	result["gptReplyDeliveryBoundaryPassed"] = true
	result["gptCommandBoundaryAndQuotaPassed"] = true
	result["authenticatedLogsPassed"] = true
	result["logsRedactionAndCursorPassed"] = true
	result["logExportPassed"] = true
	result["appearanceAssetsPassed"] = true
	result["pluginSearchAssetsPassed"] = true
	result["pluginDraftAssetsPassed"] = true
	result["connectionDraftAssetsPassed"] = true
	result["connectionDiscardAssetsPassed"] = true
	result["logIssueFilterPassed"] = true
	result["logTimeFilterPassed"] = true
	result["backupPreviewExpiryAssetsPassed"] = true
	result["backupCatalogExportAssetsPassed"] = true
	result["logSummaryPassed"] = true
	result["logQuickRangeAssetsPassed"] = true
	result["pluginSortAssetsPassed"] = true
	result["pluginDiscardOneAssetsPassed"] = true
	result["backupSortAssetsPassed"] = true
	result["backupSizeSummaryAssetsPassed"] = true
	result["logReadOnceAssetsPassed"] = true
	result["backupDateFilterAssetsPassed"] = true
	result["pluginStateCatalogAssetsPassed"] = true
	result["logCopyEntryAssetsPassed"] = true
	result["aiDraftFieldsAssetsPassed"] = true
	result["connectionDraftFieldsAssetsPassed"] = true
	result["logClearDisplayAssetsPassed"] = true
	if _, err := os.Stat(filepath.Join(runtimeRoot, "data/Chat/dict.txt")); !os.IsNotExist(err) {
		return fmt.Errorf("startup unexpectedly loaded wordcount dictionary: %v", err)
	}
	result["wordcountLazyStartupPassed"] = true
	result["adminPasswordChangePassed"] = true
	result["backupImpactPassed"] = true
	result["connectionAdditionPassed"] = true
	result["connectionRemovalPassed"] = true
	result["startupDifferencesPassed"] = true
	result["reverseConnectionPassed"] = true
	result["reverseDriverAuthAndCommandPassed"] = true
	result["backupSearchAssetsPassed"] = true
	result["backupPaginationAssetsPassed"] = true
	result["pluginPaginationAssetsPassed"] = true
	result["pluginPendingAssetsPassed"] = true
	result["pluginResponseStatusPassed"] = true
	result["pluginUsageExportAssetsPassed"] = true
	result["pluginSaveAllAssetsPassed"] = true
	result["pluginSequentialSwitchesPassed"] = true
	result["aiDraftAssetsPassed"] = true
	result["aiDiscardAssetsPassed"] = true
	result["aiPrivateConversationPassed"] = true
	result["overviewAlertsPassed"] = true
	result["backupListAndManualCreatePassed"] = true
	result["backupPreviewRestoreAndConflictPassed"] = true
	result["backupCleanupPreviewAndRetentionPassed"] = true
	result["controlPolicyBackupsPassed"] = true
	result["controlPolicyFailureFeedbackPassed"] = true
	encoded, err = json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(runtimeRoot, "result.json"), encoded, 0600); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(artifactRoot, "latest.json"), encoded, 0600); err != nil {
		return err
	}
	fmt.Println(string(encoded))
	fmt.Println("ARTIFACTS=" + runtimeRoot)
	if *uiWait {
		fmt.Println("SMOKE_ADMIN_URL=http://" + adminAddress)
		select {
		case <-time.After(5 * time.Minute):
		case <-ctx.Done():
		}
	}
	return nil
}

func prepareFixture(root, destination string) error {
	for _, relative := range []string{
		"data/Control/kanban.png", "data/Control/zbpuwu.png",
		"data/Font/GlowSansSC-Normal-ExtraBold.ttf", "data/Font/Impact.ttf",
		"data/Font/Torus-Regular.otf", "data/Font/FOT-Yuruka Std.ttf",
		"data/hyaku/小倉百人一首.csv",
		"data/Dish/dishes.db", "data/Pig/pig_data.json", "data/McFish/articlesInfo.json",
	} {
		if err := copyFixtureFile(filepath.Join(root, filepath.FromSlash(relative)), filepath.Join(destination, filepath.FromSlash(relative))); err != nil {
			return fmt.Errorf("fixture %s: %w", relative, err)
		}
	}
	return nil
}

func copyFixtureFile(source, destination string) error {
	in, err := os.Open(source)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(destination), 0700); err != nil {
		return err
	}
	out, err := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	_, err = io.Copy(out, in)
	closeErr := out.Close()
	if err != nil {
		return err
	}
	return closeErr
}

// Every event below uses synthetic IDs and the isolated OneBot connection.
// Assert replies AND subsequent plugin behavior so a success message alone
// cannot hide an unchanged response gate or a change leaking to another group.
func verifyGroupResponses(send func(int64, int64, string, string, bool) (string, error)) error {
	cases := []struct {
		name, role, command, want string
		uid, gid                  int64
		atBot                     bool
	}{
		{"owner response without mention rejected", "member", "/响应", "", 12345678, 101, false},
		{"missing mention preserves silence", "member", "翻转 abc", "", 23456789, 101, false},
		{"owner response with mention", "member", "/响应", "工作", 12345678, 101, true},
		{"current group enabled", "member", "翻转 abc", "ɔqɐ", 23456789, 101, false},
		{"other group stays silent", "member", "翻转 abc", "", 23456789, 202, false},
		{"member silence rejected", "member", "/沉默", "", 23456789, 101, false},
		{"member silence with mention rejected", "member", "/沉默", "", 23456789, 101, true},
		{"rejected silence preserves response", "member", "翻转 abc", "ɔqɐ", 23456789, 101, false},
		{"admin silence without mention rejected", "admin", "/沉默\n", "", 23456789, 101, false},
		{"missing mention preserves response", "member", "翻转 abc", "ɔqɐ", 23456789, 101, false},
		{"admin silence with mention", "admin", "/沉默\n", "休息", 23456789, 101, true},
		{"current group silent", "member", "翻转 abc", "", 23456789, 101, false},
		{"member response rejected", "member", "/响应", "", 23456789, 101, false},
		{"member response with mention rejected", "member", "/响应", "", 23456789, 101, true},
		{"response prefix rejected", "admin", "/响应xxx", "", 23456789, 101, true},
		{"response arguments rejected", "admin", "/响应 202", "", 23456789, 101, true},
		{"rejected response preserves silence", "member", "翻转 abc", "", 23456789, 101, false},
		{"admin response without mention rejected", "admin", "/响应\r\n", "", 23456789, 101, false},
		{"missing admin mention preserves silence", "member", "翻转 abc", "", 23456789, 101, false},
		{"admin response with whitespace and mention", "admin", "/响应\r\n", "工作", 23456789, 101, true},
		{"admin response enables plugin", "member", "翻转 abc", "ɔqɐ", 23456789, 101, false},
		{"group owner silence", "owner", "/沉默", "休息", 34567890, 101, true},
		{"admin response with mention", "admin", "/response", "工作", 23456789, 101, true},
		{"other group remains silent", "member", "翻转 abc", "", 23456789, 202, false},
		{"group owner silence alias", "owner", "/silence", "休息", 34567890, 101, true},
		{"global response restricted to bot owner", "admin", "/全局响应", "", 23456789, 101, true},
		{"global denial preserves other group", "member", "翻转 abc", "", 23456789, 202, false},
	}
	for _, c := range cases {
		text, err := send(c.uid, c.gid, c.role, c.command, c.atBot)
		if err != nil || (c.want == "" && text != "") || (c.want != "" && !strings.Contains(text, c.want)) {
			return fmt.Errorf("group response check %q failed: got %q, want %q, error %v", c.name, text, c.want, err)
		}
	}
	return nil
}

func smokeServiceListNodes(raw json.RawMessage) error {
	var nodes []struct {
		Type string
		Data map[string]json.RawMessage
	}
	if err := json.Unmarshal(raw, &nodes); err != nil {
		return err
	}
	if len(nodes) < 2 || len(nodes) > 21 {
		return fmt.Errorf("service list forward bounds invalid")
	}
	for i, node := range nodes {
		var sender string
		_ = json.Unmarshal(node.Data["uin"], &sender)
		if node.Type != "node" || sender != "87654321" {
			return fmt.Errorf("service list node sender/type invalid")
		}
		text, err := smokeMessageText(node.Data["content"])
		if err != nil || !utf8.ValidString(text) || utf8.RuneCountInString(text) > 1200 {
			return fmt.Errorf("service list text invalid or exceeds bounds")
		}
		if i == 0 && !strings.Contains(text, "服务列表与用法") {
			return fmt.Errorf("service list header missing")
		}
		if i > 0 && (!strings.Contains(text, "插件：") || !strings.Contains(text, "当前会话：") || !strings.Contains(text, "使用方式：")) {
			return fmt.Errorf("service list usage/state missing")
		}
	}
	return nil
}

func smokeMessageText(raw json.RawMessage) (string, error) {
	// OneBot accepts string/CQ and segment-array forms, including node content.
	if bytes.HasPrefix(bytes.TrimSpace(raw), []byte(`"`)) {
		var text string
		err := json.Unmarshal(raw, &text)
		return text, err
	}
	var segments []struct {
		Type string
		Data map[string]json.RawMessage
	}
	if err := json.Unmarshal(raw, &segments); err != nil {
		return "", err
	}
	var out strings.Builder
	for _, segment := range segments {
		switch segment.Type {
		case "text":
			var text string
			if err := json.Unmarshal(segment.Data["text"], &text); err != nil {
				return "", err
			}
			out.WriteString(text)
		case "node":
			text, err := smokeMessageText(segment.Data["content"])
			if err != nil {
				return "", err
			}
			out.WriteString(text)
		}
	}
	return out.String(), nil
}

func verifyBackupImpacts(preview map[string]any) error {
	backup, ok := preview["backup"].(map[string]any)
	if !ok {
		return fmt.Errorf("impact backup metadata missing")
	}
	scope, _ := backup["scope"].(string)
	labels := map[string][]string{
		"startup-config": {"正向 OneBot 连接", "反向 OneBot 连接", "机器人主人", "其他机器人参数", "其他启动设置"},
		"admin-options":  {"后台启用状态", "后台访问地址"},
		"control-db":     {"响应规则与AI扩展", "全局用户封禁", "插件开关继承标记", "插件与其他库内数据"},
	}[scope]
	items, ok := preview["impacts"].([]any)
	if !ok || len(labels) == 0 || len(items) < len(labels) || len(items) > len(labels)+1 {
		return fmt.Errorf("impact categories incomplete")
	}
	changed, rows := false, float64(0)
	for i, raw := range items {
		item, ok := raw.(map[string]any)
		if !ok {
			return fmt.Errorf("invalid impact record")
		}
		label := "字段名称或表示差异"
		if i < len(labels) {
			label = labels[i]
		} else if scope != "startup-config" {
			return fmt.Errorf("unexpected impact category")
		}
		if item["label"] != label {
			return fmt.Errorf("unexpected impact label")
		}
		flag, ok := item["changed"].(bool)
		if !ok {
			return fmt.Errorf("impact change marker missing")
		}
		changed = changed || flag
		for key := range item {
			switch key {
			case "label", "changed", "currentCount", "backupCount", "unit":
			default:
				return fmt.Errorf("unexpected impact value")
			}
		}
		if scope != "admin-options" && i < len(labels) {
			for _, key := range []string{"currentCount", "backupCount"} {
				count, ok := item[key].(float64)
				if !ok || count < 0 || count != float64(int(count)) {
					return fmt.Errorf("invalid impact count")
				}
				if key == "backupCount" {
					rows += count
				}
			}
		} else if item["currentCount"] != nil || item["backupCount"] != nil {
			return fmt.Errorf("unexpected impact count")
		}
	}
	if preview["changed"] != changed || scope == "control-db" && preview["rows"] != rows {
		return fmt.Errorf("impact summary inconsistent")
	}
	encoded, _ := json.Marshal(preview)
	for _, secret := range []string{"synthetic-model-key", "local-smoke-only", "127.0.0.1:3010", "127.0.0.1:3111"} {
		if bytes.Contains(encoded, []byte(secret)) {
			return fmt.Errorf("impact preview leaked a value")
		}
	}
	return nil
}

func verifyPluginResponseStatus(do func(string, string, any, string) (int, map[string]any, error), sendControl func(int64, string) (string, error), sendGroup func(int64, int64, string, string, bool) (string, error)) error {
	read := func(group string, global, local bool) error {
		status, before, err := do("GET", "/api/backups", nil, "")
		if err != nil || status != 200 {
			return fmt.Errorf("response status backup count unavailable")
		}
		status, value, err := do("GET", "/api/plugins?group="+group, nil, "")
		if err != nil || status != 200 || value["groupId"] != group {
			return fmt.Errorf("response status read failed")
		}
		state, ok := value["response"].(map[string]any)
		source := "none"
		if global {
			source = "global"
		} else if local {
			source = "group"
		}
		if !ok || len(state) != 4 || state["globalEnabled"] != global || state["groupEnabled"] != local || state["effectiveEnabled"] != (global || local) || state["source"] != source {
			return fmt.Errorf("response status disagreed with command rules")
		}
		status, after, err := do("GET", "/api/backups", nil, "")
		if err != nil || status != 200 || len(before["backups"].([]any)) != len(after["backups"].([]any)) {
			return fmt.Errorf("response status read created a backup")
		}
		return nil
	}
	for _, group := range []string{"0", "101", "202"} {
		if err := read(group, false, false); err != nil {
			return err
		}
	}
	text, err := sendGroup(12345678, 101, "member", "/响应", true)
	if err != nil || !strings.Contains(text, "工作") {
		return fmt.Errorf("response status group setup failed")
	}
	if err := read("101", false, true); err != nil {
		return err
	}
	if err := read("202", false, false); err != nil {
		return err
	}
	text, err = sendControl(12345678, "/全局响应")
	if err != nil || !strings.Contains(text, "全部位置工作") {
		return fmt.Errorf("response status global setup failed")
	}
	if err := read("0", true, false); err != nil {
		return err
	}
	if err := read("101", true, true); err != nil {
		return err
	}
	if err := read("202", true, false); err != nil {
		return err
	}
	text, err = sendGroup(12345678, 202, "member", "/沉默", true)
	if err != nil || !strings.Contains(text, "休息") {
		return fmt.Errorf("response status covered group silence failed")
	}
	if err := read("202", true, false); err != nil {
		return err
	}
	text, err = sendControl(12345678, "/全局沉默")
	if err != nil || !strings.Contains(text, "休息") {
		return fmt.Errorf("response status global cleanup failed")
	}
	if err := read("101", false, true); err != nil {
		return err
	}
	if err := read("202", false, false); err != nil {
		return err
	}
	text, err = sendGroup(12345678, 101, "member", "/沉默", true)
	if err != nil || !strings.Contains(text, "休息") {
		return fmt.Errorf("response status group cleanup failed")
	}
	return read("101", false, false)
}

func verifySequentialPluginSaves(do func(string, string, any, string) (int, map[string]any, error), csrf string) error {
	read := func() (map[string]map[string]any, error) {
		status, value, err := do("GET", "/api/plugins?group=0", nil, "")
		if err != nil || status != 200 || value["controlAvailable"] != true {
			return nil, fmt.Errorf("sequential plugin read failed")
		}
		states := map[string]map[string]any{}
		for _, raw := range value["plugins"].([]any) {
			p := raw.(map[string]any)
			if p["name"] == "chat" || p["name"] == "atri" {
				states[p["name"].(string)] = p["state"].(map[string]any)
			}
		}
		if len(states) != 2 {
			return nil, fmt.Errorf("sequential plugin fixtures absent")
		}
		return states, nil
	}
	backups := func() (int, error) {
		status, value, err := do("GET", "/api/backups", nil, "")
		if err != nil || status != 200 {
			return 0, fmt.Errorf("sequential backup count failed")
		}
		n := 0
		for _, raw := range value["backups"].([]any) {
			if raw.(map[string]any)["scope"] == "control-db" {
				n++
			}
		}
		return n, nil
	}
	write := func(name, mode string, old map[string]any, expected int) (map[string]any, error) {
		status, value, err := do("POST", "/api/plugin-switch", map[string]any{"name": name, "groupId": "0", "mode": mode, "revision": old["revision"]}, csrf)
		if err != nil || status != expected {
			return nil, fmt.Errorf("sequential plugin save status %d expected %d", status, expected)
		}
		if status != 200 {
			return nil, nil
		}
		state, ok := value["state"].(map[string]any)
		if !ok || value["name"] != name || value["groupId"] != "0" || state["globalMode"] != mode || len(fmt.Sprint(state["revision"])) != 64 {
			return nil, fmt.Errorf("sequential plugin result mismatched")
		}
		return state, nil
	}
	initial, err := read()
	if err != nil {
		return err
	}
	before, err := backups()
	if err != nil {
		return err
	}
	for _, name := range []string{"chat", "atri"} {
		mode := "enabled"
		if initial[name]["globalMode"] == mode {
			mode = "disabled"
		}
		if _, err := write(name, mode, initial[name], 200); err != nil {
			return err
		}
	}
	after, err := backups()
	if err != nil || after != before+2 {
		return fmt.Errorf("sequential changes did not create individual backups")
	}
	changed, err := read()
	if err != nil {
		return err
	}
	// The first restore succeeds; the second uses its old revision and conflicts.
	// The already successful item stays committed, matching UI stop-on-error rules.
	first, err := write("chat", initial["chat"]["globalMode"].(string), changed["chat"], 200)
	if err != nil {
		return err
	}
	if _, err := write("atri", initial["atri"]["globalMode"].(string), initial["atri"], 409); err != nil {
		return err
	}
	partial, err := read()
	if err != nil || partial["chat"]["revision"] != first["revision"] || partial["atri"]["revision"] != changed["atri"]["revision"] {
		return fmt.Errorf("partial failure lost successful or remaining state")
	}
	if _, err := write("atri", initial["atri"]["globalMode"].(string), partial["atri"], 200); err != nil {
		return err
	}
	final, err := read()
	if err != nil || final["chat"]["globalMode"] != initial["chat"]["globalMode"] || final["atri"]["globalMode"] != initial["atri"]["globalMode"] {
		return fmt.Errorf("sequential fixture cleanup did not restore modes")
	}
	after, err = backups()
	if err != nil || after != before+4 {
		return fmt.Errorf("failed item wrote a backup or restore failed")
	}
	if _, err := write("chat", final["chat"]["globalMode"].(string), final["chat"], 200); err != nil {
		return err
	}
	unchanged, err := backups()
	if err != nil || unchanged != after {
		return fmt.Errorf("unchanged plugin produced an unnecessary backup")
	}
	return nil
}

func verifyStartupSummary(value map[string]any, target string, running, saved int, expectedChanged bool) error {
	items, ok := value["startupDifferences"].([]any)
	labels := []string{"正向 OneBot 连接", "反向 OneBot 连接", "机器人主人", "其他机器人参数", "其他启动设置", "字段名称或表示差异"}
	if !ok || len(items) < 5 || len(items) > 6 {
		return fmt.Errorf("startup difference categories missing")
	}
	changed, found := false, false
	for i, raw := range items {
		item, ok := raw.(map[string]any)
		if !ok || item["label"] != labels[i] {
			return fmt.Errorf("startup difference labels incorrect")
		}
		state, ok := item["changed"].(bool)
		if !ok {
			return fmt.Errorf("startup difference state missing")
		}
		changed = changed || state
		for key := range item {
			if key != "label" && key != "changed" && key != "unit" && key != "runningCount" && key != "savedCount" {
				return fmt.Errorf("startup difference contains unapproved field")
			}
		}
		if i < 5 {
			a, aok := item["runningCount"].(float64)
			b, bok := item["savedCount"].(float64)
			if !aok || !bok || a < 0 || b < 0 {
				return fmt.Errorf("startup difference counts invalid")
			}
			if item["label"] == target {
				found = true
				if a != float64(running) || b != float64(saved) || state != expectedChanged {
					return fmt.Errorf("startup difference target did not match snapshot")
				}
			}
		} else if item["runningCount"] != nil || item["savedCount"] != nil {
			return fmt.Errorf("representation row invented counts")
		}
	}
	if !found || value["restartRequired"] != changed {
		return fmt.Errorf("startup summary disagrees with restart decision")
	}
	return nil
}

func verifyLogSummary(value map[string]any) error {
	summary, ok := value["summary"].(map[string]any)
	if !ok || len(summary) != 2 {
		return fmt.Errorf("log summary schema unavailable")
	}
	matched, ok := summary["matched"].(float64)
	retained, rok := value["retained"].(float64)
	if !ok || !rok || matched < 0 || matched > retained || retained > 1000 || matched != float64(int(matched)) {
		return fmt.Errorf("log matching total invalid")
	}
	levels, ok := summary["levels"].(map[string]any)
	if !ok || len(levels) != 8 {
		return fmt.Errorf("log level summary schema invalid")
	}
	var total float64
	for _, name := range []string{"panic", "fatal", "error", "warning", "info", "debug", "trace", "other"} {
		n, ok := levels[name].(float64)
		if !ok || n < 0 || n > 1000 || n != float64(int(n)) {
			return fmt.Errorf("log level count invalid")
		}
		total += n
	}
	if total != matched || len(value["entries"].([]any)) > int(matched) {
		return fmt.Errorf("log summary differs from retained matching rows")
	}
	return nil
}

func verifyAIConfigChecks(value map[string]any) error {
	global, ok := value["global"].(map[string]any)
	if !ok {
		return fmt.Errorf("AI checks global snapshot unavailable")
	}
	fields, _ := global["fields"].(map[string]any)
	keys, _ := global["keys"].(map[string]any)
	checks, ok := global["checks"].(map[string]any)
	if !ok || len(checks) != 3 {
		return fmt.Errorf("AI checks missing independent interfaces")
	}
	for _, spec := range []struct{ name, prefix, key string }{{"chat", "", "Key"}, {"agent", "Agent", "AgentKey"}, {"image", "Image", "ImageKey"}} {
		check, ok := checks[spec.name].(map[string]any)
		if !ok || len(check) != 8 {
			return fmt.Errorf("AI check schema invalid")
		}
		for _, value := range check {
			if _, ok := value.(bool); !ok {
				return fmt.Errorf("AI check exposes field contents")
			}
		}
		kind, _ := fields[spec.prefix+"Type"].(float64)
		model, _ := fields[spec.prefix+"ModelName"].(string)
		address, _ := fields[spec.prefix+"API"].(string)
		u, err := url.Parse(address)
		addressValid := err == nil && u.Hostname() != "" && (u.Scheme == "http" || u.Scheme == "https") && u.User == nil && u.RawQuery == "" && u.Fragment == ""
		if addressValid && u.Port() != "" {
			port, err := strconv.ParseUint(u.Port(), 10, 16)
			addressValid = err == nil && port != 0
		}
		modelConfigured, addressConfigured := strings.TrimSpace(model) != "", strings.TrimSpace(address) != ""
		keyConfigured := keys[spec.key] == true
		typeValid, modelRequired, keyRequired := kind >= 0 && kind <= 2, spec.name != "image", kind != 1
		expected := map[string]bool{"typeValid": typeValid, "modelConfigured": modelConfigured, "modelRequired": modelRequired, "addressConfigured": addressConfigured, "addressValid": addressValid, "keyConfigured": keyConfigured, "keyRequired": keyRequired,
			"complete": typeValid && addressConfigured && addressValid && (modelConfigured || !modelRequired) && (keyConfigured || !keyRequired)}
		for name, want := range expected {
			if check[name] != want {
				return fmt.Errorf("AI check differs from returned snapshot: %s/%s", spec.name, name)
			}
		}
	}
	return nil
}

func verifyAdmin(baseURL, dataDir, modelURL string, sendControl func(int64, string) (string, error), sendGroup func(int64, int64, string, string, bool) (string, error)) error {
	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar, Timeout: 15 * time.Second}
	for _, asset := range []struct{ path, kind, marker string }{
		{"/", "text/html; charset=utf-8", `data-theme-select`},
		{"/", "text/html; charset=utf-8", `id="password-change-form"`},
		{"/security.js", "text/javascript; charset=utf-8", `changeAdminPassword`},
		{"/", "text/html; charset=utf-8", `id="ai-config-checks"`},
		{"/ai.js", "text/javascript; charset=utf-8", `renderAIConfigChecks`},
		{"/style.css", "text/css; charset=utf-8", `.ai-config-check-list`},
		{"/", "text/html; charset=utf-8", `id="log-summary-state"`},
		{"/logs.js", "text/javascript; charset=utf-8", `renderLogSummary`},
		{"/", "text/html; charset=utf-8", `id="log-recent-15m"`},
		{"/", "text/html; charset=utf-8", `id="log-recent-1h"`},
		{"/", "text/html; charset=utf-8", `id="log-recent-24h"`},
		{"/", "text/html; charset=utf-8", `id="log-any-time"`},
		{"/logs.js", "text/javascript; charset=utf-8", `applyRecentLogRange`},
		{"/", "text/html; charset=utf-8", `id="plugin-sort"`},
		{"/app.js", "text/javascript; charset=utf-8", `function sortPluginItems`},
		{"/", "text/html; charset=utf-8", `id="model-request-categories"`},
		{"/overview.js", "text/javascript; charset=utf-8", `function renderModelRequests`},
		{"/", "text/html; charset=utf-8", `id="reply-task-categories"`},
		{"/reply-metrics.js", "text/javascript; charset=utf-8", `function renderReplyTasks`},
		{"/", "text/html; charset=utf-8", `id="download-reply-statistics"`},
		{"/reply-metrics.js", "text/javascript; charset=utf-8", `function downloadReplyStatistics`},
		{"/", "text/html; charset=utf-8", `id="download-model-observation"`},
		{"/model-observation.js", "text/javascript; charset=utf-8", `function downloadModelObservation`},
		{"/style.css", "text/css; charset=utf-8", `.model-observation-actions`},
		{"/style.css", "text/css; charset=utf-8", `.model-request-grid`},
		{"/app.js", "text/javascript; charset=utf-8", `function discardOnePluginDraft`},
		{"/app.js", "text/javascript; charset=utf-8", `name !== plugin.name`},
		{"/app.js", "text/javascript; charset=utf-8", `放弃此项更改`},
		{"/style.css", "text/css; charset=utf-8", `.plugin-single-draft`},
		{"/", "text/html; charset=utf-8", `id="backup-sort"`},
		{"/backups.js", "text/javascript; charset=utf-8", `function sortBackupEntries`},
		{"/", "text/html; charset=utf-8", `id="backup-size-summary"`},
		{"/backups.js", "text/javascript; charset=utf-8", `function renderBackupSizeSummary`},
		{"/style.css", "text/css; charset=utf-8", `.backup-size-scopes`},
		{"/", "text/html; charset=utf-8", `id="read-logs-once"`},
		{"/logs.js", "text/javascript; charset=utf-8", `async function readLogsOnce`},
		{"/", "text/html; charset=utf-8", `id="backup-from-date"`},
		{"/", "text/html; charset=utf-8", `id="backup-until-date"`},
		{"/backups.js", "text/javascript; charset=utf-8", `function applyBackupDateFilter`},
		{"/", "text/html; charset=utf-8", `id="download-plugin-catalog"`},
		{"/app.js", "text/javascript; charset=utf-8", `function downloadPluginCatalog`},
		{"/logs.js", "text/javascript; charset=utf-8", `async function copyLogEntry`},
		{"/style.css", "text/css; charset=utf-8", `.log-copy-actions`},
		{"/", "text/html; charset=utf-8", `id="ai-draft-fields"`},
		{"/", "text/html; charset=utf-8", `id="connection-draft-fields"`},
		{"/app.js", "text/javascript; charset=utf-8", `function renderConnectionDraftFields`},
		{"/", "text/html; charset=utf-8", `id="clear-log-display"`},
		{"/logs.js", "text/javascript; charset=utf-8", `function clearLogDisplay`},
		{"/style.css", "text/css; charset=utf-8", `.connection-draft-field-list`},
		{"/ai.js", "text/javascript; charset=utf-8", `function renderAIDraftFields`},
		{"/style.css", "text/css; charset=utf-8", `.ai-draft-field-list`},
		{"/style.css", "text/css; charset=utf-8", `.log-summary-levels`},
		{"/theme.js", "text/javascript; charset=utf-8", `zerobot.localadmin.theme.v1`},
		{"/appearance.css", "text/css; charset=utf-8", `:root[data-theme="dark"]`},
		{"/", "text/html; charset=utf-8", `id="plugin-state-filter"`},
		{"/", "text/html; charset=utf-8", `id="plugin-only-drafts"`},
		{"/", "text/html; charset=utf-8", `id="discard-plugin-drafts"`},
		{"/app.js", "text/javascript; charset=utf-8", `discardPluginDrafts`},
		{"/", "text/html; charset=utf-8", `id="save-all-plugin-drafts"`},
		{"/", "text/html; charset=utf-8", `id="stop-plugin-batch"`},
		{"/app.js", "text/javascript; charset=utf-8", `saveAllPluginDrafts`},
		{"/", "text/html; charset=utf-8", `id="plugin-response-panel"`},
		{"/app.js", "text/javascript; charset=utf-8", `renderPluginResponse`},
		{"/", "text/html; charset=utf-8", `id="copy-plugin-usage"`},
		{"/", "text/html; charset=utf-8", `id="download-plugin-usage"`},
		{"/app.js", "text/javascript; charset=utf-8", `copyPluginUsage`},
		{"/app.js", "text/javascript; charset=utf-8", `downloadPluginUsage`},
		{"/style.css", "text/css; charset=utf-8", `.usage-actions`},
		{"/style.css", "text/css; charset=utf-8", `.plugin-response-panel`},
		{"/app.js", "text/javascript; charset=utf-8", `pluginBatchResultValid`},
		{"/style.css", "text/css; charset=utf-8", `#plugin-batch-status`},
		{"/app.js", "text/javascript; charset=utf-8", `pluginOnlyDrafts`},
		{"/", "text/html; charset=utf-8", `id="plugin-page-size"`},
		{"/", "text/html; charset=utf-8", `id="plugin-page-status"`},
		{"/app.js", "text/javascript; charset=utf-8", `changePluginPage`},
		{"/app.js", "text/javascript; charset=utf-8", `renderPluginPagination`},
		{"/", "text/html; charset=utf-8", `id="clear-plugin-filter"`},
		{"/app.js", "text/javascript; charset=utf-8", `[p.name, p.brief, p.usage]`},
		{"/appearance.css", "text/css; charset=utf-8", `.plugin-filter-label select`},
		{"/", "text/html; charset=utf-8", `id="ai-draft-status"`},
		{"/", "text/html; charset=utf-8", `id="discard-ai-global"`},
		{"/ai.js", "text/javascript; charset=utf-8", `discardAIForm`},
		{"/ai.js", "text/javascript; charset=utf-8", `data-ai-discard`},
		{"/ai.js", "text/javascript; charset=utf-8", `confirmAIDiscard`},
		{"/ai.js", "text/javascript; charset=utf-8", `beforeunload`},
		{"/app.js", "text/javascript; charset=utf-8", `csrf === requestToken`},
		{"/", "text/html; charset=utf-8", `id="plugin-draft-status"`},
		{"/app.js", "text/javascript; charset=utf-8", `confirmPluginDiscard`},
		{"/app.js", "text/javascript; charset=utf-8", `generation !== pluginGeneration`},
		{"/", "text/html; charset=utf-8", `id="connection-delete-preview"`},
		{"/connections-delete.js", "text/javascript; charset=utf-8", `deletionCurrent`},
		{"/", "text/html; charset=utf-8", `id="startup-differences"`},
		{"/app.js", "text/javascript; charset=utf-8", `renderStartupDifferences`},
		{"/style.css", "text/css; charset=utf-8", `.startup-difference-row`},
		{"/", "text/html; charset=utf-8", `id="connection-add-kind"`},
		{"/connections-add.js", "text/javascript; charset=utf-8", `本机反向监听必须填写 Token`},
		{"/", "text/html; charset=utf-8", `id="connection-add-form"`},
		{"/connections-add.js", "text/javascript; charset=utf-8", `connectionAddSnapshot`},
		{"/", "text/html; charset=utf-8", `id="connection-draft-status"`},
		{"/", "text/html; charset=utf-8", `id="discard-current-connection"`},
		{"/", "text/html; charset=utf-8", `id="log-from"`},
		{"/", "text/html; charset=utf-8", `id="log-until"`},
		{"/logs.js", "text/javascript; charset=utf-8", `function appliedLogQuery`},
		{"/", "text/html; charset=utf-8", `id="discard-connection-owners"`},
		{"/app.js", "text/javascript; charset=utf-8", `discardConnectionDraft`},
		{"/app.js", "text/javascript; charset=utf-8", `settings === connectionSettings`},
		{"/", "text/html; charset=utf-8", `id="connection-draft-confirmation"`},
		{"/app.js", "text/javascript; charset=utf-8", `finishConnectionDraftConfirmation`},
		{"/app.js", "text/javascript; charset=utf-8", `confirmConnectionDiscard`},
		{"/app.js", "text/javascript; charset=utf-8", `generation !== connectionGeneration`},
		{"/backups.js", "text/javascript; charset=utf-8", `resetConnections()`},
		{"/", "text/html; charset=utf-8", `id="view-issue-logs"`},
		{"/", "text/html; charset=utf-8", `id="clear-log-filter"`},
		{"/logs.js", "text/javascript; charset=utf-8", `applyLogFilters`},
		{"/", "text/html; charset=utf-8", `id="restore-impact-list"`},
		{"/", "text/html; charset=utf-8", `id="restore-impact-status"`},
		{"/backups.js", "text/javascript; charset=utf-8", `renderRestoreImpacts`},
		{"/", "text/html; charset=utf-8", `id="backup-scope-filter"`},
		{"/", "text/html; charset=utf-8", `id="backup-state-filter"`},
		{"/", "text/html; charset=utf-8", `id="restore-backup-id"`},
		{"/backups.js", "text/javascript; charset=utf-8", `applyBackupFilter`},
		{"/backups.js", "text/javascript; charset=utf-8", `backupReadFailed`},
		{"/backups.js", "text/javascript; charset=utf-8", `尚未取得记录`},
		{"/", "text/html; charset=utf-8", `id="restore-expiry-state"`},
		{"/", "text/html; charset=utf-8", `id="cleanup-expiry-state"`},
		{"/backups.js", "text/javascript; charset=utf-8", `function renderBackupExpiry`},
		{"/backups.js", "text/javascript; charset=utf-8", `expiredBackupPreviews`},
		{"/style.css", "text/css; charset=utf-8", `.preview-expiry-state`},
		{"/", "text/html; charset=utf-8", `id="download-backup-catalog"`},
		{"/", "text/html; charset=utf-8", `id="backup-export-status"`},
		{"/backups.js", "text/javascript; charset=utf-8", `function downloadBackupCatalog`},
		{"/backups.js", "text/javascript; charset=utf-8", `zerobot-backup-catalog`},
		{"/style.css", "text/css; charset=utf-8", `.backup-export`},
		{"/", "text/html; charset=utf-8", `id="backup-page-size"`},
		{"/", "text/html; charset=utf-8", `id="backup-page-status"`},
		{"/backups.js", "text/javascript; charset=utf-8", `changeBackupPage`},
		{"/backups.js", "text/javascript; charset=utf-8", `prefers-reduced-motion`},
		{"/", "text/html; charset=utf-8", `id="ai-scope-kind"`},
		{"/ai.js", "text/javascript; charset=utf-8", `body.privateId=value.privateId`},
	} {
		rsp, err := client.Get(baseURL + asset.path)
		if err != nil {
			return err
		}
		body, readErr := io.ReadAll(io.LimitReader(rsp.Body, 128<<10))
		rsp.Body.Close()
		if readErr != nil || rsp.StatusCode != 200 || rsp.Header.Get("Content-Type") != asset.kind || rsp.Header.Get("Cache-Control") != "no-store" || !bytes.Contains(body, []byte(asset.marker)) {
			return fmt.Errorf("admin asset check failed: %s", asset.path)
		}
	}
	do := func(method, path string, body any, csrf string) (int, map[string]any, error) {
		var encoded []byte
		if body != nil {
			encoded, _ = json.Marshal(body)
		}
		req, err := http.NewRequest(method, baseURL+path, bytes.NewReader(encoded))
		if err != nil {
			return 0, nil, err
		}
		if method == "POST" {
			req.Header.Set("Origin", baseURL)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Local-Request", "1")
			req.Header.Set("X-CSRF-Token", csrf)
		}
		rsp, err := client.Do(req)
		if err != nil {
			return 0, nil, err
		}
		defer rsp.Body.Close()
		var value map[string]any
		err = json.NewDecoder(rsp.Body).Decode(&value)
		if err == nil && path == "/api/backups/preview" && rsp.StatusCode == 200 {
			err = verifyBackupImpacts(value)
		}
		return rsp.StatusCode, value, err
	}
	status, _, err := do("GET", "/api/plugins", nil, "")
	if err != nil || status != 401 {
		return fmt.Errorf("unauthed admin request was not denied: %d %v", status, err)
	}
	status, _, err = do("GET", "/api/logs", nil, "")
	if err != nil || status != 401 {
		return fmt.Errorf("unauthed logs were not denied")
	}
	status, _, err = do("GET", "/api/logs/export", nil, "")
	if err != nil || status != 401 {
		return fmt.Errorf("unauthed log export was not denied")
	}
	status, _, err = do("GET", "/api/backups", nil, "")
	if err != nil || status != 401 {
		return fmt.Errorf("unauthed backups were not denied")
	}
	code, err := os.ReadFile(filepath.Join(dataDir, "setup-code.txt"))
	if err != nil {
		return err
	}
	status, session, err := do("POST", "/api/setup", map[string]string{"setupCode": strings.TrimSpace(string(code)), "password": "Local-ui-test-password-only"}, "")
	if err != nil || status != 200 {
		return fmt.Errorf("isolated admin setup failed: %d %v", status, err)
	}
	csrf, _ := session["csrfToken"].(string)
	// Only this run's isolated options file is edited; runtime uses explicit CLI flags.
	optionsPath := filepath.Join(dataDir, "options.json")
	optionsBefore := []byte("{\"enabled\":true,\"address\":\"127.0.0.1:3010\"}")
	if err = os.WriteFile(optionsPath, optionsBefore, 0600); err != nil {
		return err
	}
	status, optionsBackup, err := do("POST", "/api/backups", map[string]string{"scope": "admin-options"}, csrf)
	if err != nil || status != 200 || optionsBackup["backupId"] == "" {
		return fmt.Errorf("options backup fixture failed: %d %v", status, err)
	}
	status, equalOptions, err := do("POST", "/api/backups/preview", map[string]any{"id": optionsBackup["backupId"]}, csrf)
	if err != nil || status != 200 || equalOptions["changed"] != false {
		return fmt.Errorf("equal options impacts failed: %d %v", status, err)
	}
	optionsChanged := []byte("{\"enabled\":false,\"address\":\"127.0.0.1:3111\"}")
	if err = os.WriteFile(optionsPath, optionsChanged, 0600); err != nil {
		return err
	}
	status, changedOptions, err := do("POST", "/api/backups/preview", map[string]any{"id": optionsBackup["backupId"]}, csrf)
	if err != nil || status != 200 || changedOptions["changed"] != true || changedOptions["restartRequired"] != true {
		return fmt.Errorf("changed options impacts failed: %d %v", status, err)
	}
	optionsAfter, readErr := os.ReadFile(optionsPath)
	if readErr != nil || !bytes.Equal(optionsAfter, optionsChanged) {
		return fmt.Errorf("preview changed isolated options")
	}

	status, overview, err := do("GET", "/api/overview", nil, "")
	if err != nil || status != 200 || overview["connectionState"] != "connected" {
		return fmt.Errorf("isolated runtime overview failed: %d %v", status, err)
	}
	modelRequests, statsOK := overview["modelRequests"].(map[string]any)
	if !statsOK || modelRequests["started"] != float64(0) || modelRequests["active"] != float64(0) || modelRequests["completed"] != float64(0) || modelRequests["failed"] != float64(0) || modelRequests["incomplete"] != float64(0) || modelRequests["startedAt"] == nil || modelRequests["checkedAt"] == nil {
		return fmt.Errorf("isolated zero model statistics absent or incorrect")
	}
	categories, statsOK := modelRequests["categories"].([]any)
	if !statsOK || len(categories) != 4 {
		return fmt.Errorf("model statistics category bounds absent")
	}
	accounts, _ := overview["botAccounts"].([]any)
	if len(accounts) != 1 || accounts[0] != "87654321" {
		return fmt.Errorf("overview does not reflect actual isolated bot")
	}
	alerts, ok := overview["alerts"].(map[string]any)
	if !ok {
		return fmt.Errorf("overview retained alerts absent")
	}
	alertRows, ok := alerts["entries"].([]any)
	if !ok || len(alertRows) > 5 || alerts["warningCount"] == nil || alerts["errorCount"] == nil || alerts["retained"] == nil {
		return fmt.Errorf("overview alert bounds/counts absent")
	}
	lastID := uint64(^uint64(0))
	for _, item := range alertRows {
		row, ok := item.(map[string]any)
		if !ok {
			return fmt.Errorf("invalid overview alert row")
		}
		id, _ := strconv.ParseUint(fmt.Sprint(row["id"]), 10, 64)
		level := fmt.Sprint(row["level"])
		if id >= lastID || (level != "warning" && level != "error" && level != "fatal" && level != "panic") {
			return fmt.Errorf("overview alerts not ordered/filtered")
		}
		lastID = id
	}
	assetResponse, assetErr := client.Get(baseURL + "/overview.js")
	if assetErr != nil {
		return assetErr
	}
	assetBody, assetErr := io.ReadAll(assetResponse.Body)
	assetResponse.Body.Close()
	if assetErr != nil || assetResponse.StatusCode != 200 || !strings.Contains(string(assetBody), "visibilitychange") || !strings.Contains(string(assetBody), "AbortController") {
		return fmt.Errorf("overview polling asset unavailable")
	}
	status, list, err := do("GET", "/api/plugins", nil, "")
	if err != nil || status != 200 {
		return fmt.Errorf("isolated plugin registry failed")
	}
	plugins, _ := list["plugins"].([]any)
	found := false
	for _, value := range plugins {
		if p, ok := value.(map[string]any); ok && p["name"] == "chat" {
			found = true
		}
	}
	if !found || float64(len(plugins)) != overview["pluginCount"] {
		return fmt.Errorf("actual plugin registry does not match overview")
	}
	pages := (len(plugins) + 19) / 20
	for page := 1; page <= pages; page++ {
		text, err := sendControl(23456789, fmt.Sprintf("/服务列表 %d", page))
		if err != nil || !strings.Contains(text, fmt.Sprintf("第 %d/%d 页", page, pages)) {
			return fmt.Errorf("service list page %d failed: %v", page, err)
		}
		for _, item := range plugins[(page-1)*20 : min(page*20, len(plugins))] {
			plugin := item.(map[string]any)
			if !strings.Contains(text, "插件："+plugin["name"].(string)+"\n") {
				return fmt.Errorf("service list lost registered plugin")
			}
		}
	}
	for _, arg := range []string{"0", "abc", strconv.Itoa(pages + 1)} {
		text, err := sendControl(23456789, "/服务列表 "+arg)
		if err != nil || !strings.Contains(text, "页") {
			return fmt.Errorf("service list invalid page not rejected")
		}
	}
	if text, err := sendControl(23456789, "/service_list"); err != nil || !strings.Contains(text, "第 1/") {
		return fmt.Errorf("service list alias failed")
	}
	if text, err := sendGroup(23456789, 202, "member", "/服务列表", false); err != nil || text != "" {
		return fmt.Errorf("service list lost group @ condition")
	}
	if text, err := sendGroup(23456789, 202, "member", "/服务列表", true); err != nil || !strings.Contains(text, "当前群") {
		return fmt.Errorf("group service list forward failed")
	}
	if err := verifySequentialPluginSaves(do, csrf); err != nil {
		return err
	}
	// Re-read versions after the fixture writes, even when effective modes match.
	status, list, err = do("GET", "/api/plugins", nil, "")
	if err != nil || status != 200 {
		return fmt.Errorf("plugin versions after sequential fixture unavailable")
	}
	plugins = list["plugins"].([]any)
	// Isolate routing checks from unsolicited random chat and AI responses.
	// These overrides affect only this synthetic database. GPT remains enabled.
	for _, item := range plugins {
		p := item.(map[string]any)
		if p["name"] == "aichat" || p["name"] == "chat" || p["name"] == "atri" || p["name"] == "thesaurus" {
			state := p["state"].(map[string]any)
			status, _, err = do("POST", "/api/plugin-switch", map[string]any{"name": p["name"], "groupId": "0", "mode": "disabled", "revision": state["revision"]}, csrf)
			if err != nil || status != 200 {
				return fmt.Errorf("isolated unsolicited-reply disable failed: %v status %d: %v", p["name"], status, err)
			}
		}
	}
	if err := verifyPluginResponseStatus(do, sendControl, sendGroup); err != nil {
		return err
	}
	pluginState := func(group string) (map[string]any, error) {
		status, list, err := do("GET", "/api/plugins?group="+group, nil, "")
		if err != nil || status != 200 || list["controlAvailable"] != true {
			return nil, fmt.Errorf("plugin state read failed")
		}
		items, _ := list["plugins"].([]any)
		for _, item := range items {
			if p, ok := item.(map[string]any); ok && p["name"] == "chrev" {
				state, _ := p["state"].(map[string]any)
				if state != nil {
					return state, nil
				}
			}
		}
		return nil, fmt.Errorf("compiled chrev control missing")
	}
	setPlugin := func(group, mode string) (map[string]any, error) {
		state, err := pluginState(group)
		if err != nil {
			return nil, err
		}
		body := map[string]any{"name": "chrev", "groupId": group, "mode": mode, "revision": state["revision"]}
		status, _, err := do("POST", "/api/plugin-switch", body, "invalid")
		if err != nil || status != 403 {
			return nil, fmt.Errorf("plugin save lacked CSRF protection")
		}
		status, value, err := do("POST", "/api/plugin-switch", body, csrf)
		if err != nil || status != 200 {
			return nil, fmt.Errorf("plugin save failed: %d %v", status, err)
		}
		saved, _ := value["state"].(map[string]any)
		return saved, nil
	}
	if text, err := sendControl(12345678, "/响应"); err != nil || !strings.Contains(text, "工作") {
		return fmt.Errorf("synthetic response permission setup failed: %v", err)
	}
	if _, err := setPlugin("0", "disabled"); err != nil {
		return err
	}
	if text, err := sendControl(12345678, "翻转 abc"); err != nil || text != "" {
		return fmt.Errorf("disabled plugin still replied: %v", err)
	}
	if text, err := sendControl(23456789, "/全局启用 chrev"); err != nil || text != "" {
		return fmt.Errorf("non-owner global command was accepted: %v", err)
	}
	state, err := pluginState("101")
	if err != nil || state["effectiveEnabled"] != false || state["source"] != "global" {
		return fmt.Errorf("disabled global state lost")
	}
	if text, err := sendControl(12345678, "/全局启用 chrev"); err != nil || !strings.Contains(text, "已全局启用") {
		return fmt.Errorf("owner global command failed: %v", err)
	}
	state, err = pluginState("101")
	if err != nil || state["effectiveEnabled"] != true || state["source"] != "global" {
		return fmt.Errorf("chat update not visible to admin")
	}
	if text, err := sendControl(12345678, "翻转 abc"); err != nil || text != "ɔqɐ" {
		return fmt.Errorf("enabled plugin did not reply correctly: %v", err)
	}
	state, err = setPlugin("101", "disabled")
	if err != nil || state["effectiveEnabled"] != true || state["source"] != "global" {
		return fmt.Errorf("group did not respect global override")
	}
	if _, err = setPlugin("0", "inherit"); err != nil {
		return err
	}
	state, err = pluginState("101")
	if err != nil || state["effectiveEnabled"] != false || state["source"] != "group" {
		return fmt.Errorf("inherited global did not reveal group override")
	}
	state, err = pluginState("202")
	if err != nil || state["effectiveEnabled"] != true || state["source"] != "default" {
		return fmt.Errorf("group override leaked")
	}
	if _, err = setPlugin("101", "inherit"); err != nil {
		return err
	}
	status, beforeAIRead, err := do("GET", "/api/backups", nil, "")
	if err != nil || status != 200 {
		return fmt.Errorf("AI checks backup baseline unavailable")
	}
	status, ai, err := do("GET", "/api/ai?group=101", nil, "")
	if err != nil || status != 200 || ai["available"] != true {
		return fmt.Errorf("AI configuration unavailable")
	}
	if err := verifyAIConfigChecks(ai); err != nil {
		return err
	}
	status, afterAIRead, err := do("GET", "/api/backups", nil, "")
	if err != nil || status != 200 || len(beforeAIRead["backups"].([]any)) != len(afterAIRead["backups"].([]any)) {
		return fmt.Errorf("AI checks read created a backup")
	}
	aiGlobal := ai["global"].(map[string]any)
	aiChange := map[string]any{"revision": aiGlobal["revision"], "fields": map[string]any{"API": modelURL, "ModelName": "smoke-web-model", "Type": 0, "SystemP": "local mock system"}, "keys": map[string]any{"Key": map[string]any{"action": "replace", "value": "synthetic-model-key"}}}
	status, _, err = do("POST", "/api/ai", aiChange, "bad")
	if err != nil || status != 403 {
		return fmt.Errorf("AI save lacked CSRF protection")
	}
	status, ai, err = do("POST", "/api/ai", aiChange, csrf)
	if err != nil || status != 200 {
		return fmt.Errorf("AI web save failed")
	}
	if err := verifyAIConfigChecks(ai); err != nil || ai["global"].(map[string]any)["checks"].(map[string]any)["chat"].(map[string]any)["complete"] != true {
		return fmt.Errorf("AI checks did not update after saving complete fields: %v", err)
	}
	encodedAI, _ := json.Marshal(ai)
	if bytes.Contains(encodedAI, []byte("synthetic-model-key")) {
		return fmt.Errorf("AI API leaked a key")
	}
	status, _, err = do("POST", "/api/ai", aiChange, csrf)
	if err != nil || status != 409 {
		return fmt.Errorf("AI stale web version accepted")
	}
	if text, err := sendControl(12345678, "查看AI聊天配置"); err != nil || !strings.Contains(text, "smoke-web-model") || strings.Contains(text, "synthetic-model-key") {
		return fmt.Errorf("AI web config invisible/secret exposed in chat")
	}
	if text, err := sendControl(12345678, "设置AI聊天模型名smoke-chat-model"); err != nil || !strings.Contains(text, "成功") {
		return fmt.Errorf("AI chat save failed")
	}
	status, ai, err = do("GET", "/api/ai?group=101", nil, "")
	if err != nil || status != 200 || ai["global"].(map[string]any)["fields"].(map[string]any)["ModelName"] != "smoke-chat-model" {
		return fmt.Errorf("AI chat update invisible to web")
	}
	if err := verifyAIConfigChecks(ai); err != nil {
		return err
	}
	for _, item := range ai["groups"].([]any) {
		group := item.(map[string]any)
		change := map[string]any{"service": group["service"], "groupId": "101", "revision": group["revision"], "temperature": 82}
		if group["service"] == "aichatcfg" {
			change["rate"] = 11
			change["useAgent"] = false
			change["voiceOutput"] = false
			change["respondAt"] = true
		}
		status, _, err = do("POST", "/api/ai-group", change, csrf)
		if err != nil || status != 200 {
			return fmt.Errorf("AI group save failed")
		}
	}
	if text, err := sendGroup(12345678, 101, "member", "/响应", true); err != nil || !strings.Contains(text, "工作") {
		return fmt.Errorf("AI group response setup failed")
	}
	if text, err := sendGroup(12345678, 101, "member", "设置AI聊天温度80", true); err != nil || !strings.Contains(text, "成功") {
		return fmt.Errorf("AI chat group setting failed")
	}
	status, ai, err = do("GET", "/api/ai?group=101", nil, "")
	if err != nil || status != 200 {
		return fmt.Errorf("AI group reread failed")
	}
	group := ai["groups"].([]any)[0].(map[string]any)
	if group["temperature"] != float64(80) || group["rate"] != float64(11) {
		return fmt.Errorf("AI group chat/web state differs")
	}
	status, other, err := do("GET", "/api/ai?group=202", nil, "")
	if err != nil || status != 200 || other["groups"].([]any)[0].(map[string]any)["temperature"] != float64(0) {
		return fmt.Errorf("AI group settings leaked to other group")
	}
	status, privateAI, err := do("GET", "/api/ai?private=12345678", nil, "")
	if err != nil || status != 200 {
		return fmt.Errorf("private AI configuration unavailable")
	}
	var privateChatChange map[string]any
	for _, item := range privateAI["groups"].([]any) {
		conversation := item.(map[string]any)
		if conversation["privateId"] != "12345678" || conversation["groupId"] != "" {
			return fmt.Errorf("private AI target identity is ambiguous")
		}
		change := map[string]any{"service": conversation["service"], "privateId": "12345678", "revision": conversation["revision"], "temperature": 81}
		if conversation["service"] == "aichatcfg" {
			change["rate"], change["useAgent"], change["voiceOutput"], change["respondAt"] = 0, false, false, false
			privateChatChange = change
		}
		status, _, err = do("POST", "/api/ai-group", change, csrf)
		if err != nil || status != 200 {
			return fmt.Errorf("private AI save failed")
		}
	}
	if text, err := sendControl(12345678, "设置AI聊天温度76"); err != nil || !strings.Contains(text, "成功") {
		return fmt.Errorf("private AI chat configuration failed")
	}
	status, privateAI, err = do("GET", "/api/ai?private=12345678", nil, "")
	if err != nil || status != 200 || privateAI["groups"].([]any)[0].(map[string]any)["temperature"] != float64(76) || privateAI["groups"].([]any)[1].(map[string]any)["temperature"] != float64(81) {
		return fmt.Errorf("private AI chat/web storage mismatch")
	}
	status, _, err = do("POST", "/api/ai-group", privateChatChange, csrf)
	if err != nil || status != 409 {
		return fmt.Errorf("stale private AI update accepted")
	}
	for _, path := range []string{"/api/ai?group=12345678", "/api/ai?private=12345679"} {
		status, isolated, err := do("GET", path, nil, "")
		if err != nil || status != 200 || isolated["groups"].([]any)[0].(map[string]any)["temperature"] != float64(0) {
			return fmt.Errorf("private AI change leaked: %s", path)
		}
	}
	status, _, err = do("GET", "/api/ai?group=12345678&private=12345678", nil, "")
	if err != nil || status != 400 {
		return fmt.Errorf("ambiguous private/group target accepted")
	}
	if text, err := sendGroup(12345678, 101, "member", "/gpt group smoke", false); err != nil || !strings.Contains(text, "本地模型模拟回复") {
		return fmt.Errorf("group GPT request failed: %v %q", err, text)
	}
	if text, err := sendControl(12345678, "/gpt quoted smoke"); err != nil || !strings.Contains(text, "本地模型模拟回复") {
		return fmt.Errorf("private quoted GPT request failed: %v %q", err, text)
	}
	if text, err := sendControl(12345678, "/gpt second"); err != nil || !strings.Contains(text, "30 秒") {
		return fmt.Errorf("GPT request quota not enforced")
	}
	for _, command := range []string{"/gptabc", "普通提及 /gpt"} {
		if text, err := sendControl(12345678, command); err != nil || text != "" {
			return fmt.Errorf("GPT command boundary accepted %q", command)
		}
	}
	if text, err := sendControl(12345678, "/gpt"); err != nil || !strings.Contains(text, "填写问题") {
		return fmt.Errorf("empty GPT did not show usage")
	}
	if err := verifyAIRolePermissions(do, sendControl, sendGroup, csrf, modelURL); err != nil {
		return err
	}
	if err := verifyVoiceCommands(do, csrf, sendControl, sendGroup); err != nil {
		return err
	}
	if text, err := sendGroup(12345678, 101, "member", "/沉默", true); err != nil || !strings.Contains(text, "休息") {
		return fmt.Errorf("AI fixture response cleanup failed")
	}
	status, logs, err := do("GET", "/api/logs?limit=200", nil, "")
	if err != nil || status != 200 || len(logs["entries"].([]any)) == 0 {
		return fmt.Errorf("runtime logs unavailable")
	}
	if err := verifyLogSummary(logs); err != nil {
		return err
	}
	logJSON, _ := json.Marshal(logs)
	if bytes.Contains(logJSON, []byte("synthetic-model-key")) || bytes.Contains(logJSON, []byte("local-smoke-only")) || !bytes.Contains(logJSON, []byte("正文已隐藏")) {
		return fmt.Errorf("runtime logs did not protect credentials/message bodies")
	}
	status, updatedOverview, err := do("GET", "/api/overview", nil, "")
	if err != nil || status != 200 || updatedOverview["alerts"] == nil {
		return fmt.Errorf("post-AI overview alerts unavailable")
	}
	stats, ok := updatedOverview["modelRequests"].(map[string]any)
	if !ok || stats["started"] != float64(11) || stats["completed"] != float64(11) || stats["active"] != float64(0) || stats["failed"] != float64(0) || stats["incomplete"] != float64(0) {
		return fmt.Errorf("post-AI model statistics disagree with 11 synthetic requests")
	}
	rows, ok := stats["categories"].([]any)
	if !ok || len(rows) != 4 {
		return fmt.Errorf("post-AI model categories missing")
	}
	for i, want := range []float64{4, 2, 5, 0} {
		row, ok := rows[i].(map[string]any)
		if !ok || row["started"] != want || row["completed"] != want || row["active"] != float64(0) {
			return fmt.Errorf("post-AI model category differs: %d", i)
		}
	}
	statsJSON, _ := json.Marshal(stats)
	for _, secret := range []string{"synthetic-model-key", "smoke-chat-model", modelURL, "quoted smoke", "本地模型模拟回复"} {
		if bytes.Contains(statsJSON, []byte(secret)) {
			return fmt.Errorf("model aggregates retained sensitive request data")
		}
	}
	alertJSON, _ := json.Marshal(updatedOverview["alerts"])
	if bytes.Contains(alertJSON, []byte("synthetic-model-key")) || bytes.Contains(alertJSON, []byte("local-smoke-only")) {
		return fmt.Errorf("overview alerts exposed credentials")
	}
	for _, path := range []string{"/api/logs/export", "/api/logs/export?level=info&q=Websocket"} {
		rsp, err := client.Get(baseURL + path)
		if err != nil {
			return err
		}
		body, readErr := io.ReadAll(io.LimitReader(rsp.Body, (3<<20)+1))
		rsp.Body.Close()
		if readErr != nil || rsp.StatusCode != 200 || len(body) > 3<<20 || rsp.Header.Get("Cache-Control") != "no-store" || rsp.Header.Get("Content-Type") != "text/plain; charset=utf-8" || !strings.HasPrefix(rsp.Header.Get("Content-Disposition"), `attachment; filename="zerobot-logs-`) {
			return fmt.Errorf("runtime log export/download headers failed")
		}
		if bytes.Contains(body, []byte("synthetic-model-key")) || bytes.Contains(body, []byte("local-smoke-only")) || !bytes.Contains(body, []byte("[INFO]")) {
			return fmt.Errorf("runtime export redaction failed")
		}
		if path == "/api/logs/export" && !bytes.Contains(body, []byte("正文已隐藏")) {
			return fmt.Errorf("export exposed message bodies")
		}
		if path != "/api/logs/export" {
			for _, line := range strings.Split(string(body), "\n") {
				if strings.HasPrefix(line, "[") && (!strings.Contains(line, "[INFO]") || !strings.Contains(line, "Websocket")) {
					return fmt.Errorf("export filter mismatch")
				}
			}
		}
	}
	status, filtered, err := do("GET", "/api/logs?level=info&q=Websocket", nil, "")
	if err != nil || status != 200 || len(filtered["entries"].([]any)) == 0 {
		return fmt.Errorf("runtime log filters failed")
	}
	if err := verifyLogSummary(filtered); err != nil {
		return err
	}
	for _, item := range filtered["entries"].([]any) {
		row := item.(map[string]any)
		if row["level"] != "info" || !strings.Contains(row["message"].(string), "Websocket") {
			return fmt.Errorf("runtime log filter mismatch")
		}
	}
	// Exercise inclusive time filtering and export against actual process timestamps.
	timeRow := filtered["entries"].([]any)[0].(map[string]any)
	timeQuery := url.Values{"from": {timeRow["time"].(string)}, "until": {timeRow["time"].(string)}, "level": {"info"}, "q": {"Websocket"}}.Encode()
	status, timed, err := do("GET", "/api/logs?"+timeQuery, nil, "")
	if err != nil || status != 200 || len(timed["entries"].([]any)) == 0 {
		return fmt.Errorf("runtime time range failed")
	}
	if err := verifyLogSummary(timed); err != nil {
		return err
	}
	status, timedTail, err := do("GET", "/api/logs?"+timeQuery+"&after="+timed["cursor"].(string)+"&instance="+timed["instance"].(string)+"&limit=1", nil, "")
	if err != nil || status != 200 || len(timedTail["entries"].([]any)) != 0 {
		return fmt.Errorf("log summary incremental time fixture failed")
	}
	if err := verifyLogSummary(timedTail); err != nil {
		return err
	}
	beforeSummary, _ := json.Marshal(timed["summary"])
	afterSummary, _ := json.Marshal(timedTail["summary"])
	if !bytes.Equal(beforeSummary, afterSummary) {
		return fmt.Errorf("log summary depended on incremental cursor/limit")
	}
	status, hiddenSummary, err := do("GET", "/api/logs?q=synthetic-model-key", nil, "")
	if err != nil || status != 200 || hiddenSummary["summary"].(map[string]any)["matched"] != float64(0) {
		return fmt.Errorf("log summary inferred a raw model key")
	}
	for _, item := range timed["entries"].([]any) {
		if item.(map[string]any)["time"] != timeRow["time"] {
			return fmt.Errorf("runtime time range boundary mismatch")
		}
	}
	timeResponse, err := client.Get(baseURL + "/api/logs/export?" + timeQuery)
	if err != nil {
		return err
	}
	timeBody, timeErr := io.ReadAll(io.LimitReader(timeResponse.Body, (3<<20)+1))
	timeResponse.Body.Close()
	if timeErr != nil || timeResponse.StatusCode != 200 || !bytes.Contains(timeBody, []byte("时间范围（包含边界，UTC）")) || !bytes.Contains(timeBody, []byte("#"+timeRow["id"].(string)+" ")) {
		return fmt.Errorf("runtime time export failed")
	}
	for _, endpoint := range []string{"/api/logs", "/api/logs/export"} {
		status, _, err = do("GET", endpoint+"?from=bad-time", nil, "")
		if err != nil || status != 400 {
			return fmt.Errorf("runtime invalid time not rejected")
		}
		status, _, err = do("GET", endpoint+"?from=2026-10-09T08:00:01Z&until=2026-10-09T08:00:00Z", nil, "")
		if err != nil || status != 400 {
			return fmt.Errorf("runtime reversed time not rejected")
		}
	}
	status, issues, err := do("GET", "/api/logs?level=issues", nil, "")
	if err != nil || status != 200 {
		return fmt.Errorf("runtime issues filter unavailable")
	}
	if err := verifyLogSummary(issues); err != nil {
		return err
	}
	for _, item := range issues["entries"].([]any) {
		row := item.(map[string]any)
		switch row["level"] {
		case "warning", "error", "fatal", "panic":
		default:
			return fmt.Errorf("runtime issues filter included a normal level")
		}
	}
	rsp, err := client.Get(baseURL + "/api/logs/export?level=issues")
	if err != nil {
		return err
	}
	issueBody, readIssueErr := io.ReadAll(io.LimitReader(rsp.Body, (3<<20)+1))
	rsp.Body.Close()
	if readIssueErr != nil || rsp.StatusCode != 200 || rsp.Header.Get("Content-Type") != "text/plain; charset=utf-8" || rsp.Header.Get("Cache-Control") != "no-store" || bytes.Contains(issueBody, []byte("synthetic-model-key")) || bytes.Contains(issueBody, []byte("local-smoke-only")) {
		return fmt.Errorf("runtime issues export failed")
	}
	for _, line := range strings.Split(string(issueBody), "\n") {
		if strings.HasPrefix(line, "[") && !strings.Contains(line, "[WARNING]") && !strings.Contains(line, "[ERROR]") && !strings.Contains(line, "[FATAL]") && !strings.Contains(line, "[PANIC]") {
			return fmt.Errorf("runtime issues export included a normal level")
		}
	}
	status, tail, err := do("GET", "/api/logs?after="+logs["cursor"].(string)+"&instance="+logs["instance"].(string), nil, "")
	if err != nil || status != 200 || tail["reset"] != false || tail["capacity"] != float64(1000) {
		return fmt.Errorf("runtime log cursor failed")
	}
	if err := verifyLogSummary(tail); err != nil {
		return err
	}
	status, settings, err := do("GET", "/api/connections", nil, "")
	if err != nil || status != 200 || settings["available"] != true || settings["restartRequired"] != false {
		return fmt.Errorf("isolated connection settings read failed")
	}
	if err := verifyStartupSummary(settings, "正向 OneBot 连接", 1, 1, false); err != nil {
		return err
	}
	settingsJSON, _ := json.Marshal(settings)
	if bytes.Contains(settingsJSON, []byte("local-smoke-only")) {
		return fmt.Errorf("connection settings leaked token")
	}
	revision, _ := settings["revision"].(string)
	change := map[string]any{"revision": revision, "superUsers": []string{"12345678", "23456789"}}
	status, _, err = do("POST", "/api/connections", change, "invalid")
	if err != nil || status != 403 {
		return fmt.Errorf("connection save lacked CSRF protection")
	}
	status, saved, err := do("POST", "/api/connections", change, csrf)
	if err != nil || status != 200 || saved["restartRequired"] != true || saved["backupId"] == "" {
		return fmt.Errorf("isolated connection save/backup failed")
	}
	if err := verifyStartupSummary(saved, "机器人主人", 1, 2, true); err != nil {
		return err
	}
	status, _, err = do("POST", "/api/connections", change, csrf)
	if err != nil || status != 409 {
		return fmt.Errorf("stale connection save was accepted")
	}
	// Restore a synthetic startup backup and confirm running configuration stays intact.
	restore := func(id string) error {
		status, preview, err := do("POST", "/api/backups/preview", map[string]string{"id": id}, csrf)
		if err != nil || status != 200 || preview["changed"] != true {
			return fmt.Errorf("restore preview failed: %d %v", status, err)
		}
		body := map[string]any{"id": id, "revision": preview["revision"], "expiresAt": preview["expiresAt"], "token": preview["token"], "confirm": true}
		status, _, err = do("POST", "/api/backups/restore", body, "invalid")
		if err != nil || status != 403 {
			return fmt.Errorf("restore lacked CSRF protection")
		}
		status, result, err := do("POST", "/api/backups/restore", body, csrf)
		if err != nil || status != 200 || result["changed"] != true || result["backupId"] == "" {
			return fmt.Errorf("restore commit failed: %d %v", status, err)
		}
		status, _, err = do("POST", "/api/backups/restore", body, csrf)
		if err != nil || status != 409 {
			return fmt.Errorf("stale restore ticket accepted")
		}
		return nil
	}
	if err = restore(saved["backupId"].(string)); err != nil {
		return err
	}
	status, settings, err = do("GET", "/api/connections", nil, "")
	if err != nil || status != 200 || settings["restartRequired"] != false || len(settings["superUsers"].([]any)) != 1 {
		return fmt.Errorf("startup restore did not recover saved config")
	}
	// Detect a same-count Token-only change without altering the running driver.
	if err := verifyStartupSummary(settings, "机器人主人", 1, 1, false); err != nil {
		return err
	}
	secretChange := map[string]any{"revision": settings["revision"], "connection": map[string]any{"kind": "ws", "index": 0, "tokenAction": "replace", "token": "startup-summary-only-private"}}
	status, secretSaved, err := do("POST", "/api/connections", secretChange, csrf)
	if err != nil || status != 200 || secretSaved["restartRequired"] != true {
		return fmt.Errorf("startup summary secret-only save failed")
	}
	if err := verifyStartupSummary(secretSaved, "正向 OneBot 连接", 1, 1, true); err != nil {
		return err
	}
	secretJSON, _ := json.Marshal(secretSaved)
	if bytes.Contains(secretJSON, []byte("startup-summary-only-private")) {
		return fmt.Errorf("startup summary leaked secret")
	}
	if err = restore(secretSaved["backupId"].(string)); err != nil {
		return err
	}
	status, settings, err = do("GET", "/api/connections", nil, "")
	if err != nil || status != 200 {
		return fmt.Errorf("startup summary restore read failed")
	}
	if err := verifyStartupSummary(settings, "正向 OneBot 连接", 1, 1, false); err != nil {
		return err
	}
	// Add only an isolated forward driver, then restore the pre-addition fixture.
	addition := map[string]any{"revision": settings["revision"], "addConnection": map[string]any{"url": "ws://127.0.0.1:17777/new", "token": "synthetic-addition-key", "withoutToken": false}}
	status, _, err = do("POST", "/api/connections", addition, "invalid")
	if err != nil || status != 403 {
		return fmt.Errorf("addition lacked CSRF protection")
	}
	status, added, err := do("POST", "/api/connections", addition, csrf)
	if err != nil || status != 200 || added["restartRequired"] != true || added["backupId"] == "" || len(added["connections"].([]any)) != 2 || len(added["runningConnections"].([]any)) != 1 {
		return fmt.Errorf("isolated addition failed: %d %v", status, err)
	}
	if err := verifyStartupSummary(added, "正向 OneBot 连接", 1, 2, true); err != nil {
		return err
	}
	addedJSON, _ := json.Marshal(added)
	if bytes.Contains(addedJSON, []byte("synthetic-addition-key")) {
		return fmt.Errorf("addition exposed token")
	}
	status, _, err = do("POST", "/api/connections", addition, csrf)
	if err != nil || status != 409 {
		return fmt.Errorf("stale addition accepted")
	}
	duplicate := map[string]any{"revision": added["revision"], "addConnection": map[string]any{"url": "ws://127.0.0.1:17777/new", "withoutToken": true}}
	status, _, err = do("POST", "/api/connections", duplicate, csrf)
	if err != nil || status != 400 {
		return fmt.Errorf("duplicate addition accepted")
	}
	// Remove only the added synthetic driver, preserving the active fake connection.
	removal := map[string]any{"revision": added["revision"], "removeConnection": map[string]any{"kind": "ws", "index": 1, "confirm": true}}
	status, _, err = do("POST", "/api/connections", removal, "invalid")
	if err != nil || status != 403 {
		return fmt.Errorf("removal lacked CSRF protection")
	}
	unconfirmed := map[string]any{"revision": added["revision"], "removeConnection": map[string]any{"kind": "ws", "index": 1, "confirm": false}}
	status, _, err = do("POST", "/api/connections", unconfirmed, csrf)
	if err != nil || status != 400 {
		return fmt.Errorf("removal accepted without confirmation")
	}
	status, removed, err := do("POST", "/api/connections", removal, csrf)
	if err != nil || status != 200 || removed["backupId"] == "" || removed["restartRequired"] != false || len(removed["connections"].([]any)) != 1 || len(removed["runningConnections"].([]any)) != 1 {
		return fmt.Errorf("isolated removal failed: %d %v", status, err)
	}
	if err := verifyStartupSummary(removed, "正向 OneBot 连接", 1, 1, false); err != nil {
		return err
	}
	status, _, err = do("POST", "/api/connections", removal, csrf)
	if err != nil || status != 409 {
		return fmt.Errorf("stale removal accepted")
	}
	last := map[string]any{"revision": removed["revision"], "removeConnection": map[string]any{"kind": "ws", "index": 0, "confirm": true}}
	status, _, err = do("POST", "/api/connections", last, csrf)
	if err != nil || status != 400 {
		return fmt.Errorf("last connection removal accepted")
	}
	if err = restore(removed["backupId"].(string)); err != nil {
		return err
	}
	if err = restore(added["backupId"].(string)); err != nil {
		return err
	}
	status, settings, err = do("GET", "/api/connections", nil, "")
	if err != nil || status != 200 || settings["restartRequired"] != false || len(settings["connections"].([]any)) != 1 {
		return fmt.Errorf("addition restore did not preserve running fixture")
	}
	// Reverse addition is saved only in this run's startup file; no listener starts here.
	backendURL, _ := url.Parse(baseURL)
	blockedReverse := map[string]any{"revision": settings["revision"], "addConnection": map[string]any{"kind": "wss", "url": "ws://" + backendURL.Host, "token": "synthetic-reverse-api-key"}}
	status, _, err = do("POST", "/api/connections", blockedReverse, csrf)
	if err != nil || status != 400 {
		return fmt.Errorf("reverse backend port conflict accepted")
	}
	reverseAdd := map[string]any{"revision": settings["revision"], "addConnection": map[string]any{"kind": "wss", "url": "ws://127.0.0.1:17778", "token": "synthetic-reverse-api-key"}}
	status, reverseSaved, err := do("POST", "/api/connections", reverseAdd, csrf)
	if err != nil || status != 200 || reverseSaved["backupId"] == "" || reverseSaved["restartRequired"] != true || len(reverseSaved["connections"].([]any)) != 2 || len(reverseSaved["runningConnections"].([]any)) != 1 {
		return fmt.Errorf("reverse addition API failed: %d %v", status, err)
	}
	if err := verifyStartupSummary(reverseSaved, "反向 OneBot 连接", 0, 1, true); err != nil {
		return err
	}
	reverseJSON, _ := json.Marshal(reverseSaved)
	if bytes.Contains(reverseJSON, []byte("synthetic-reverse-api-key")) {
		return fmt.Errorf("reverse addition leaked token")
	}
	if err = restore(reverseSaved["backupId"].(string)); err != nil {
		return err
	}
	status, manual, err := do("POST", "/api/backups", map[string]string{"scope": "control-db"}, csrf)
	if err != nil || status != 200 || manual["backupId"] == "" {
		return fmt.Errorf("manual control snapshot failed")
	}
	if _, err = setPlugin("0", "disabled"); err != nil {
		return err
	}
	status, ai, err = do("GET", "/api/ai", nil, "")
	if err != nil || status != 200 {
		return fmt.Errorf("AI restore fixture read failed")
	}
	currentModel := ai["global"].(map[string]any)["fields"].(map[string]any)["ModelName"]
	status, _, err = do("POST", "/api/ai", map[string]any{"revision": ai["global"].(map[string]any)["revision"], "fields": map[string]any{"ModelName": "restore-temporary-model"}}, csrf)
	if err != nil || status != 200 {
		return fmt.Errorf("AI restore fixture write failed")
	}
	if err = restore(manual["backupId"].(string)); err != nil {
		return err
	}
	state, err = pluginState("101")
	if err != nil || state["effectiveEnabled"] != true {
		return fmt.Errorf("control restore left stale plugin cache")
	}
	status, ai, err = do("GET", "/api/ai", nil, "")
	if err != nil || status != 200 || ai["global"].(map[string]any)["fields"].(map[string]any)["ModelName"] != currentModel {
		return fmt.Errorf("control restore did not recover AI settings")
	}
	if text, err := sendControl(12345678, "翻转 abc"); err != nil || text != "ɔqɐ" {
		return fmt.Errorf("restored plugin did not reply: %v", err)
	}
	status, backups, err := do("GET", "/api/backups", nil, "")
	if err != nil || status != 200 || len(backups["backups"].([]any)) < 3 {
		return fmt.Errorf("backup list unavailable")
	}
	backupJSON, _ := json.Marshal(backups)
	if bytes.Contains(backupJSON, []byte("synthetic-model-key")) || bytes.Contains(backupJSON, []byte("local-smoke-only")) {
		return fmt.Errorf("backup metadata leaked credentials")
	}
	// Age only snapshots created inside this run's isolated backup directory.
	// Preserve nested RawMessage values and large integers/checksums.
	for i := 0; i < 4; i++ {
		status, created, err := do("POST", "/api/backups", map[string]string{"scope": "startup-config"}, csrf)
		if err != nil || status != 200 {
			return fmt.Errorf("cleanup fixture creation failed")
		}
		id, ok := created["backupId"].(string)
		if !ok || filepath.Base(id) != id || !strings.HasSuffix(id, ".json") {
			return fmt.Errorf("cleanup fixture ID invalid")
		}
		path := filepath.Join(dataDir, "backups", id)
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		var fields map[string]json.RawMessage
		if json.Unmarshal(raw, &fields) != nil {
			return fmt.Errorf("cleanup fixture invalid")
		}
		fields["createdAt"], _ = json.Marshal(time.Now().UTC().AddDate(0, 0, -30-i))
		raw, _ = json.Marshal(fields)
		if os.WriteFile(path, raw, 0600) != nil {
			return fmt.Errorf("cleanup fixture aging failed")
		}
	}
	policy := map[string]int{"keepPerScope": 3, "keepDays": 7}
	status, preview, err := do("POST", "/api/backups/cleanup-preview", policy, csrf)
	if err != nil || status != 200 || len(preview["candidates"].([]any)) == 0 {
		return fmt.Errorf("cleanup preview did not select old fixtures")
	}
	cleanupBody := map[string]any{"policy": preview["policy"], "revision": preview["revision"], "asOf": preview["asOf"], "expiresAt": preview["expiresAt"], "token": preview["token"], "confirm": false}
	status, _, err = do("POST", "/api/backups/cleanup", cleanupBody, csrf)
	if err != nil || status != 400 {
		return fmt.Errorf("cleanup lacked explicit confirmation")
	}
	cleanupBody["confirm"] = true
	status, _, err = do("POST", "/api/backups/cleanup", cleanupBody, "invalid")
	if err != nil || status != 403 {
		return fmt.Errorf("cleanup lacked CSRF protection")
	}
	startupPath := filepath.Join(filepath.Dir(dataDir), "config.json")
	startupBefore, err := os.ReadFile(startupPath)
	if err != nil {
		return err
	}
	status, cleaned, err := do("POST", "/api/backups/cleanup", cleanupBody, csrf)
	if err != nil || status != 200 || cleaned["partial"] != false || len(cleaned["deleted"].([]any)) != len(preview["candidates"].([]any)) {
		return fmt.Errorf("isolated cleanup failed: %d %v", status, err)
	}
	startupAfter, err := os.ReadFile(startupPath)
	if err != nil || !bytes.Equal(startupBefore, startupAfter) {
		return fmt.Errorf("cleanup changed current startup configuration")
	}
	status, _, err = do("POST", "/api/backups/cleanup", cleanupBody, csrf)
	if err != nil || status != 409 {
		return fmt.Errorf("cleanup replay was accepted")
	}
	status, remaining, err := do("GET", "/api/backups", nil, "")
	if err != nil || status != 200 {
		return fmt.Errorf("cleanup list refresh failed")
	}
	validStartup := 0
	for _, entry := range remaining["backups"].([]any) {
		b := entry.(map[string]any)
		if b["scope"] == "startup-config" && b["valid"] == true {
			validStartup++
		}
	}
	if validStartup < 3 {
		return fmt.Errorf("cleanup lost protected backups")
	}
	if err = verifyPolicyCommands(dataDir, sendControl, sendGroup); err != nil {
		return err
	}
	// Password changes operate only on this run's synthetic credential directory.
	const originalPassword, replacementPassword = "Local-ui-test-password-only", "Local-ui-changed-password-only"
	passwordBody := func(current, next string) map[string]string {
		return map[string]string{"currentPassword": current, "newPassword": next, "confirmPassword": next}
	}
	primaryClient := client
	secondJar, _ := cookiejar.New(nil)
	client = &http.Client{Jar: secondJar, Timeout: 15 * time.Second}
	status, secondSession, err := do("POST", "/api/login", map[string]string{"password": originalPassword}, "")
	if err != nil || status != 200 {
		return fmt.Errorf("second synthetic admin login failed")
	}
	secondCSRF := secondSession["csrfToken"].(string)
	secondClient := client
	client = primaryClient
	status, _, err = do("POST", "/api/password", passwordBody("wrong", replacementPassword), csrf)
	if err != nil || status != 403 {
		return fmt.Errorf("wrong current password accepted")
	}
	status, _, err = do("GET", "/api/overview", nil, "")
	if err != nil || status != 200 {
		return fmt.Errorf("wrong password killed current session")
	}
	status, _, err = do("POST", "/api/password", passwordBody(originalPassword, replacementPassword), "invalid-csrf")
	if err != nil || status != 403 {
		return fmt.Errorf("password change skipped CSRF")
	}
	status, changed, err := do("POST", "/api/password", passwordBody(originalPassword, replacementPassword), csrf)
	if err != nil || status != 200 || changed["passwordChanged"] != true || changed["authenticated"] != false {
		return fmt.Errorf("password change failed")
	}
	status, _, err = do("GET", "/api/overview", nil, "")
	if err != nil || status != 401 {
		return fmt.Errorf("primary password session remained active")
	}
	client = secondClient
	status, _, err = do("POST", "/api/password", passwordBody(originalPassword, "another-test-password-only"), secondCSRF)
	if err != nil || status != 401 {
		return fmt.Errorf("secondary old password session remained active")
	}
	client = primaryClient
	status, _, err = do("POST", "/api/login", map[string]string{"password": originalPassword}, "")
	if err != nil || status != 401 {
		return fmt.Errorf("old password login remained valid")
	}
	status, changedSession, err := do("POST", "/api/login", map[string]string{"password": replacementPassword}, "")
	if err != nil || status != 200 {
		return fmt.Errorf("new password login failed")
	}
	csrf = changedSession["csrfToken"].(string)
	credentialBytes, err := os.ReadFile(filepath.Join(dataDir, "credential.json"))
	if err != nil || bytes.Contains(credentialBytes, []byte(originalPassword)) || bytes.Contains(credentialBytes, []byte(replacementPassword)) {
		return fmt.Errorf("password credential persistence wrong")
	}
	status, _, err = do("POST", "/api/password", passwordBody(replacementPassword, originalPassword), csrf)
	if err != nil || status != 200 {
		return fmt.Errorf("synthetic password restore failed")
	}
	status, restoredSession, err := do("POST", "/api/login", map[string]string{"password": originalPassword}, "")
	if err != nil || status != 200 {
		return fmt.Errorf("restored synthetic password login failed")
	}
	csrf = restoredSession["csrfToken"].(string)
	status, _, err = do("POST", "/api/logout", map[string]any{}, csrf)
	if err != nil || status != 200 {
		return fmt.Errorf("isolated logout failed")
	}
	status, _, err = do("GET", "/api/overview", nil, "")
	if err != nil || status != 401 {
		return fmt.Errorf("logout left admin access enabled")
	}
	status, _, err = do("GET", "/api/logs", nil, "")
	if err != nil || status != 401 {
		return fmt.Errorf("logout left log access enabled")
	}
	status, _, err = do("GET", "/api/logs/export", nil, "")
	if err != nil || status != 401 {
		return fmt.Errorf("logout left log export enabled")
	}
	status, _, err = do("GET", "/api/backups", nil, "")
	if err != nil || status != 401 {
		return fmt.Errorf("logout left backup access enabled")
	}
	for _, path := range []string{"/api/backups/cleanup-preview", "/api/backups/cleanup"} {
		status, _, err = do("POST", path, map[string]any{}, csrf)
		if err != nil || status != 401 {
			return fmt.Errorf("logout left cleanup access enabled")
		}
	}
	return nil
}

// Failure injection affects only this run's synthetic admin backup directory.
// Check snapshots, truthful command feedback, and later message routing together.
func verifyVoiceCommands(do func(string, string, any, string) (int, map[string]any, error), csrf string, private func(int64, string) (string, error), group func(int64, int64, string, string, bool) (string, error)) error {
	if text, err := private(12345678, "发送AI语音preconfig"); err != nil || !strings.Contains(text, "未选择") {
		return fmt.Errorf("unconfigured voice did not stop: %q %v", text, err)
	}
	for _, c := range []struct{ cmd, want string }{
		{"设置AI语音群号0", "正数"},
		{"设置AI语音群号303", "为303"},
		{"设置AI语音模型", "请选择语音模型序号"},
		{"0", "已选择语音模型: Synthetic voice"},
		{"查看AI语音配置", "synthetic-voice-1"},
		{"发送AI语音   ", "填写"},
		{"发送AI语音private-ok", "仅支持群聊"},
		{"发送AI语音private-failure", "仅支持群聊"},
	} {
		text, err := private(12345678, c.cmd)
		if err != nil || !strings.Contains(text, c.want) {
			return fmt.Errorf("voice private command %s failed: %q %v", c.cmd, text, err)
		}
	}
	if text, err := group(12345678, 303, "member", "/响应", true); err != nil || !strings.Contains(text, "工作") {
		return fmt.Errorf("voice group setup failed")
	}
	for _, c := range []struct{ role, cmd, want string }{
		{"member", "发送AI语音group-ok", ""},
		{"admin", "发送AI语音group-ok", "VOICE_OK"},
		{"owner", "发送AI语音group-zero-id", "VOICE_OK"},
		{"admin", "发送AI语音group-failure", "未自动重试"},
	} {
		text, err := group(23456789, 303, c.role, c.cmd, true)
		if err != nil || (c.want == "" && text != "") || (c.want != "" && !strings.Contains(text, c.want)) {
			return fmt.Errorf("voice group command %s failed: %q %v", c.cmd, text, err)
		}
	}
	if text, err := group(12345678, 404, "member", "/响应", true); err != nil || !strings.Contains(text, "工作") {
		return fmt.Errorf("second voice group setup failed")
	}
	if text, err := group(12345678, 404, "member", "发送AI语音current-group", true); err != nil || text != "VOICE_OK" {
		return fmt.Errorf("voice escaped current group: %q %v", text, err)
	}
	if err := verifyChatVoiceScope(do, csrf, private, group); err != nil {
		return err
	}
	for _, gid := range []int64{303, 404} {
		if text, err := group(12345678, gid, "member", "/沉默", true); err != nil || !strings.Contains(text, "休息") {
			return fmt.Errorf("voice group cleanup failed")
		}
	}
	return nil
}

func verifyChatVoiceScope(do func(string, string, any, string) (int, map[string]any, error), csrf string, private func(int64, string) (string, error), group func(int64, int64, string, string, bool) (string, error)) error {
	status, snapshot, err := do("GET", "/api/ai", nil, "")
	if err != nil || status != 200 {
		return fmt.Errorf("voice model snapshot failed")
	}
	global := snapshot["global"].(map[string]any)
	oldModel := global["fields"].(map[string]any)["ModelName"]
	status, _, err = do("POST", "/api/ai", map[string]any{"revision": global["revision"], "fields": map[string]any{"ModelName": "smoke-voice-chat"}}, csrf)
	if err != nil || status != 200 {
		return fmt.Errorf("voice model fixture setup failed")
	}
	switchChat := func(mode string) error {
		status, list, err := do("GET", "/api/plugins", nil, "")
		if err != nil || status != 200 {
			return fmt.Errorf("voice chat plugin read failed")
		}
		for _, item := range list["plugins"].([]any) {
			p := item.(map[string]any)
			if p["name"] != "aichat" {
				continue
			}
			state := p["state"].(map[string]any)
			status, _, err = do("POST", "/api/plugin-switch", map[string]any{"name": "aichat", "groupId": "0", "mode": mode, "revision": state["revision"]}, csrf)
			if err != nil || status != 200 {
				return fmt.Errorf("voice chat plugin switch failed")
			}
			return nil
		}
		return fmt.Errorf("voice chat plugin missing")
	}
	for _, target := range []string{"group=404", "private=12345678"} {
		status, value, err := do("GET", "/api/ai?"+target, nil, "")
		if err != nil || status != 200 {
			return fmt.Errorf("voice chat settings read failed")
		}
		for _, item := range value["groups"].([]any) {
			setting := item.(map[string]any)
			if setting["service"] != "aichatcfg" {
				continue
			}
			change := map[string]any{"service": "aichatcfg", "revision": setting["revision"], "rate": 0, "temperature": 50, "useAgent": false, "respondAt": true, "voiceOutput": true}
			if strings.HasPrefix(target, "group=") {
				change["groupId"] = "404"
			} else {
				change["privateId"] = "12345678"
			}
			status, _, err = do("POST", "/api/ai-group", change, csrf)
			if err != nil || status != 200 {
				return fmt.Errorf("voice chat settings write failed")
			}
		}
	}
	if text, err := private(12345678, "/全局响应"); err != nil || !strings.Contains(text, "工作") {
		return fmt.Errorf("voice chat response setup failed")
	}
	if err := switchChat("enabled"); err != nil {
		return err
	}
	if text, err := group(12345678, 404, "member", "VOICE_CHAT_GROUP", true); err != nil || text != "VOICE_OK" {
		return fmt.Errorf("ordinary group chat voice failed: %q %v", text, err)
	}
	if text, err := private(12345678, "VOICE_CHAT_PRIVATE"); err != nil || text != "VOICE_TEXT_OK" {
		return fmt.Errorf("ordinary private chat did not preserve text: %q %v", text, err)
	}
	if err := switchChat("disabled"); err != nil {
		return err
	}
	if text, err := private(12345678, "/全局沉默"); err != nil || !strings.Contains(text, "休息") {
		return fmt.Errorf("voice chat response cleanup failed")
	}
	status, snapshot, err = do("GET", "/api/ai", nil, "")
	if err != nil || status != 200 {
		return fmt.Errorf("voice model cleanup snapshot failed")
	}
	global = snapshot["global"].(map[string]any)
	status, _, err = do("POST", "/api/ai", map[string]any{"revision": global["revision"], "fields": map[string]any{"ModelName": oldModel}}, csrf)
	if err != nil || status != 200 {
		return fmt.Errorf("voice model fixture cleanup failed")
	}
	return nil
}

func verifyAIRolePermissions(do func(string, string, any, string) (int, map[string]any, error), private func(int64, string) (string, error), send func(int64, int64, string, string, bool) (string, error), csrf, modelURL string) error {
	switchAgent := func(mode string) error {
		status, list, err := do("GET", "/api/plugins", nil, "")
		if err != nil || status != 200 {
			return fmt.Errorf("role plugin snapshot failed")
		}
		for _, item := range list["plugins"].([]any) {
			p := item.(map[string]any)
			if p["name"] != "aichat" {
				continue
			}
			state := p["state"].(map[string]any)
			status, _, err = do("POST", "/api/plugin-switch", map[string]any{"name": "aichat", "groupId": "0", "mode": mode, "revision": state["revision"]}, csrf)
			if err != nil || status != 200 {
				return fmt.Errorf("role agent plugin switch failed")
			}
			return nil
		}
		return fmt.Errorf("role agent plugin missing")
	}
	read := func(target string) (map[string]any, error) {
		status, value, err := do("GET", "/api/ai?"+target, nil, "")
		if err != nil || status != 200 {
			return nil, fmt.Errorf("role snapshot failed: %v", err)
		}
		return value, nil
	}
	configure := func(target string, useAgent bool) error {
		value, err := read(target)
		if err != nil {
			return err
		}
		for _, item := range value["groups"].([]any) {
			c := item.(map[string]any)
			change := map[string]any{"service": c["service"], "revision": c["revision"], "temperature": 50}
			if strings.HasPrefix(target, "group=") {
				change["groupId"] = strings.TrimPrefix(target, "group=")
			} else {
				change["privateId"] = strings.TrimPrefix(target, "private=")
			}
			if c["service"] == "aichatcfg" {
				change["rate"], change["useAgent"], change["respondAt"], change["voiceOutput"] = 0, useAgent, useAgent, false
			}
			status, _, err := do("POST", "/api/ai-group", change, csrf)
			if err != nil || status != 200 {
				return fmt.Errorf("role conversation setup failed: %v", err)
			}
		}
		return nil
	}
	if err := configure("group=303", false); err != nil {
		return err
	}
	if err := configure("private=23456789", false); err != nil {
		return err
	}
	if text, err := private(12345678, "/全局响应"); err != nil || !strings.Contains(text, "工作") {
		return fmt.Errorf("role response setup failed")
	}
	for _, c := range []struct {
		name, role string
		uid, gid   int64
		allow      bool
	}{
		{"member", "member", 23456789, 303, false},
		{"admin", "admin", 23456789, 303, true},
		{"group owner", "owner", 34567890, 303, true},
		{"private non-owner", "member", 23456789, 0, false},
	} {
		target := "group=303"
		if c.gid == 0 {
			target = "private=23456789"
		}
		before, err := read(target)
		if err != nil {
			return err
		}
		text, err := send(c.uid, c.gid, c.role, "设置AI聊天温度65", c.gid > 0)
		if err != nil || (c.allow && !strings.Contains(text, "成功")) || (!c.allow && text != "") {
			return fmt.Errorf("role temperature command failed: %s %q %v", c.name, text, err)
		}
		after, err := read(target)
		if err != nil {
			return err
		}
		beforeJSON, _ := json.Marshal(before)
		afterJSON, _ := json.Marshal(after)
		if !c.allow && !bytes.Equal(beforeJSON, afterJSON) {
			return fmt.Errorf("denied setting changed state: %s", c.name)
		}
		if c.allow && after["groups"].([]any)[0].(map[string]any)["temperature"] != float64(65) {
			return fmt.Errorf("allowed setting did not persist: %s", c.name)
		}
		for _, command := range []string{"设置AI聊天模型名forbidden-role-model", "设置AI聊天不使用Agent模式", "重置AI聊天Agent", "查看AI聊天配置"} {
			if text, err = send(c.uid, c.gid, c.role, command, c.gid > 0); err != nil || text != "" {
				return fmt.Errorf("owner-only command accepted: %s %s %q %v", c.name, command, text, err)
			}
		}
		final, err := read(target)
		if err != nil {
			return err
		}
		finalJSON, _ := json.Marshal(final)
		if !bytes.Equal(afterJSON, finalJSON) {
			return fmt.Errorf("owner-only rejection changed config: %s", c.name)
		}
	}
	for _, c := range []struct {
		gid int64
		cmd string
	}{{303, "/gpt role member"}, {0, "/gpt role private"}} {
		text, err := send(23456789, c.gid, "member", c.cmd, false)
		if err != nil || !strings.Contains(text, "本地模型模拟回复") {
			return fmt.Errorf("non-owner GPT rejected: %v %q", err, text)
		}
	}
	value, err := read("group=303")
	if err != nil {
		return err
	}
	global := value["global"].(map[string]any)
	status, _, err := do("POST", "/api/ai", map[string]any{"revision": global["revision"], "fields": map[string]any{"AgentAPI": modelURL, "AgentModelName": "smoke-role-agent", "AgentType": 0}, "keys": map[string]any{"AgentKey": map[string]any{"action": "replace", "value": "synthetic-model-key"}}}, csrf)
	if err != nil || status != 200 {
		return fmt.Errorf("role agent setup failed: %v", err)
	}
	if err := configure("group=303", true); err != nil {
		return err
	}
	if err := switchAgent("enabled"); err != nil {
		return err
	}
	for _, c := range []struct {
		uid             int64
		role, cmd, want string
	}{
		{23456789, "member", "ROLE_AGENT_CHECK", "ROLE_AGENT_user"},
		{23456789, "admin", "ROLE_AGENT_CHECK", "ROLE_AGENT_admin"},
		{34567890, "owner", "ROLE_AGENT_CHECK", "ROLE_AGENT_admin"},
		{23456789, "member", "ROLE_AGENT_CROSS_INFO", "ROLE_BLOCKED_OK"},
		{23456789, "member", "ROLE_AGENT_CQ_ALL", "ROLE_BLOCKED_OK"},
	} {
		text, err := send(c.uid, 303, c.role, c.cmd, true)
		if err != nil || text != c.want {
			return fmt.Errorf("actual Agent role/action failed: %s %s got %q want %q error %v", c.role, c.cmd, text, c.want, err)
		}
		// Allow the message hook to record the API response before updating roles.
		time.Sleep(100 * time.Millisecond)
	}
	if err := configure("group=303", false); err != nil {
		return err
	}
	if err := switchAgent("disabled"); err != nil {
		return err
	}
	value, err = read("group=303")
	if err != nil {
		return err
	}
	global = value["global"].(map[string]any)
	status, _, err = do("POST", "/api/ai", map[string]any{"revision": global["revision"], "fields": map[string]any{"AgentAPI": "", "AgentModelName": ""}, "keys": map[string]any{"AgentKey": map[string]any{"action": "clear", "confirm": true}}}, csrf)
	if err != nil || status != 200 {
		return fmt.Errorf("role agent model cleanup failed")
	}
	if text, err := private(12345678, "/全局沉默"); err != nil || !strings.Contains(text, "休息") {
		return fmt.Errorf("role response cleanup failed")
	}
	return nil
}

func verifyPolicyCommands(dataDir string, private func(int64, string) (string, error), group func(int64, int64, string, string, bool) (string, error)) error {
	const owner, user = int64(12345678), int64(23456789)
	backups := filepath.Join(dataDir, "backups")
	count := func() (int, error) {
		entries, err := os.ReadDir(backups)
		if err != nil {
			return 0, err
		}
		n := 0
		for _, e := range entries {
			if e.IsDir() && strings.HasSuffix(e.Name(), "-control") {
				n++
			}
		}
		return n, nil
	}
	command := func(gid int64, cmd, want string, delta int) error {
		before, err := count()
		if err != nil {
			return err
		}
		var text string
		if gid == 0 {
			text, err = private(owner, cmd)
		} else {
			text, err = group(owner, gid, "member", cmd, true)
		}
		if err != nil || !strings.Contains(text, want) {
			return fmt.Errorf("policy command %s feedback failed: %v", cmd, err)
		}
		after, err := count()
		if err != nil {
			return err
		}
		if after-before != delta {
			return fmt.Errorf("policy command %s made %d snapshots, expected %d", cmd, after-before, delta)
		}
		return nil
	}
	route := func(uid, gid int64, want string) error {
		text, err := group(uid, gid, "member", "翻转 abc", false)
		if err != nil || text != want {
			return fmt.Errorf("policy gate for synthetic user %d/group %d failed: %v", uid, gid, err)
		}
		return nil
	}
	if err := command(101, "/响应", "工作", 1); err != nil {
		return err
	}
	if err := command(101, "/响应", "已处于响应", 0); err != nil {
		return err
	}
	for _, pair := range []struct {
		gid                  int64
		add, remove, yes, no string
	}{
		{101, "/禁止 chrev 23456789", "/允许 chrev 23456789", "已禁止", "已允许"},
		{0, "/全局禁止 chrev 23456789", "/全局允许 chrev 23456789", "已禁止", "已允许"},
		{0, "/封禁 23456789", "/解封 23456789", "已封禁", "已解封"},
	} {
		if err := route(user, 101, "ɔqɐ"); err != nil {
			return err
		}
		if err := command(pair.gid, pair.add, pair.yes, 1); err != nil {
			return err
		}
		if err := command(pair.gid, pair.add, pair.yes, 0); err != nil {
			return err
		}
		if err := route(user, 101, ""); err != nil {
			return err
		}
		if err := command(pair.gid, pair.remove, pair.no, 1); err != nil {
			return err
		}
		if err := command(pair.gid, pair.remove, pair.no, 0); err != nil {
			return err
		}
		if err := route(user, 101, "ɔqɐ"); err != nil {
			return err
		}
	}
	before, err := count()
	if err != nil {
		return err
	}
	if text, err := group(user, 101, "member", "/禁止 chrev 12345678", true); err != nil || text != "" {
		return fmt.Errorf("unauthorized policy command accepted")
	}
	after, err := count()
	if err != nil || after != before {
		return fmt.Errorf("unauthorized policy command created a backup")
	}
	for _, c := range []struct {
		gid       int64
		cmd, want string
	}{
		{101, "/禁止 chrev 34567890", "已禁止"},
		{0, "/全局禁止 chrev 45678901", "已禁止"},
		{0, "/封禁 56789012", "已封禁"},
	} {
		if err := command(c.gid, c.cmd, c.want, 1); err != nil {
			return err
		}
	}
	// The directory is inside a fresh smoke fixture, never the live admin path.
	withBlockedBackups := func(check func() error) (result error) {
		held := backups + "-held"
		if err := os.Rename(backups, held); err != nil {
			return err
		}
		defer func() {
			if err := os.Remove(backups); err != nil && !os.IsNotExist(err) {
				result = err
				return
			}
			if err := os.Rename(held, backups); err != nil {
				result = err
			}
		}()
		if err := os.WriteFile(backups, []byte("synthetic blocked backup directory"), 0600); err != nil {
			return err
		}
		return check()
	}
	failure := func(gid int64, cmd string) error {
		var text string
		var err error
		if gid == 0 {
			text, err = private(owner, cmd)
		} else {
			text, err = group(owner, gid, "member", cmd, true)
		}
		if err != nil || !strings.Contains(text, "原状态保留") || strings.Contains(text, "已禁止") || strings.Contains(text, "已允许") || strings.Contains(text, "已封禁") || strings.Contains(text, "已解封") || strings.Contains(text, "SQL") || strings.Contains(text, dataDir) {
			return fmt.Errorf("policy failure feedback leaked details or claimed success for %s", cmd)
		}
		return nil
	}
	if err := withBlockedBackups(func() error {
		for _, c := range []struct {
			gid int64
			cmd string
		}{
			{202, "/响应"}, {101, "/沉默"}, {0, "/全局响应"},
			{101, "/禁止 chrev 23456789"}, {101, "/允许 chrev 34567890"},
			{0, "/全局禁止 chrev 23456789"}, {0, "/全局允许 chrev 45678901"},
			{0, "/封禁 23456789"}, {0, "/解封 56789012"},
		} {
			if err := failure(c.gid, c.cmd); err != nil {
				return err
			}
		}
		for _, c := range []struct {
			uid, gid int64
			want     string
		}{
			{user, 101, "ɔqɐ"}, {user, 202, ""}, {34567890, 101, ""}, {45678901, 101, ""}, {56789012, 101, ""},
		} {
			if err := route(c.uid, c.gid, c.want); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}
	for _, c := range []struct {
		gid       int64
		cmd, want string
	}{
		{101, "/允许 chrev 34567890", "已允许"}, {0, "/全局允许 chrev 45678901", "已允许"},
		{0, "/解封 56789012", "已解封"}, {101, "/沉默", "休息"}, {0, "/全局响应", "全部位置工作"},
	} {
		if err := command(c.gid, c.cmd, c.want, 1); err != nil {
			return err
		}
	}
	if err := withBlockedBackups(func() error {
		if err := failure(0, "/全局沉默"); err != nil {
			return err
		}
		return route(user, 202, "ɔqɐ")
	}); err != nil {
		return err
	}
	if err := command(0, "/全局沉默", "休息", 1); err != nil {
		return err
	}
	return route(user, 101, "")
}

// A second, disposable process exercises the actual reverse driver. It has no
// daily credentials, model settings or connection to NapCat / real QQ.
func verifyReverseDriver(root, parent, binary string) error {
	fixture, err := os.MkdirTemp(parent, "reverse-driver-")
	if err != nil {
		return err
	}
	if err = prepareFixture(root, fixture); err != nil {
		return err
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	address := listener.Addr().String()
	adminListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		listener.Close()
		return err
	}
	adminAddress := adminListener.Addr().String()
	listener.Close()
	adminListener.Close()
	target := "ws://" + address
	token := "synthetic-reverse-driver-key"
	config := map[string]any{"zero": map[string]any{"nickname": []string{"七叶叶"}, "command_prefix": "/", "super_users": []int64{12345678}, "ring_len": 4096, "latency": 233000000, "max_process_time": 240000000000, "mark_message": true}, "ws": []any{}, "wss": []map[string]string{{"URL": target, "AccessToken": token}}, "force_base64_file": false}
	raw, _ := json.Marshal(config)
	configPath := filepath.Join(fixture, "config.json")
	if err = os.WriteFile(configPath, raw, 0600); err != nil {
		return err
	}
	dataDir := filepath.Join(fixture, "admin")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, filepath.Join(root, ".local", "bin", binary), "-c", configPath, "-admin", "-admin-addr", adminAddress, "-admin-data", dataDir)
	cmd.Dir = fixture
	cmd.Env = append(os.Environ(), "DEBUG_MODE=1")
	stdout, err := os.Create(filepath.Join(fixture, "stdout.log"))
	if err != nil {
		return err
	}
	defer stdout.Close()
	stderr, err := os.Create(filepath.Join(fixture, "stderr.log"))
	if err != nil {
		return err
	}
	defer stderr.Close()
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if err = cmd.Start(); err != nil {
		return err
	}
	exited := make(chan error, 1)
	go func() { exited <- cmd.Wait() }()
	defer func() {
		cancel()
		select {
		case <-exited:
		case <-time.After(3 * time.Second):
		}
	}()
	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar, Timeout: time.Second}
	base := "http://" + adminAddress
	ready := false
	for attempt := 0; attempt < 450; attempt++ {
		rsp, e := client.Get(base + "/api/session")
		if e == nil {
			io.Copy(io.Discard, rsp.Body)
			rsp.Body.Close()
			if rsp.StatusCode == 200 {
				ready = true
				break
			}
		}
		select {
		case e := <-exited:
			return fmt.Errorf("reverse fixture exited before admin ready: %v", e)
		case <-ctx.Done():
			return fmt.Errorf("reverse fixture startup timeout")
		default:
		}
		time.Sleep(200 * time.Millisecond)
	}
	if !ready {
		return fmt.Errorf("reverse fixture admin unavailable")
	}
	for _, check := range []struct {
		auth   string
		status int
	}{{"", 401}, {"Bearer incorrect-fixture-key", 403}} {
		req, _ := http.NewRequest("GET", "http://"+address+"/", nil)
		if check.auth != "" {
			req.Header.Set("Authorization", check.auth)
		}
		rsp, e := client.Do(req)
		if e != nil {
			return fmt.Errorf("reverse authentication probe unavailable: %w", e)
		}
		io.Copy(io.Discard, rsp.Body)
		rsp.Body.Close()
		if rsp.StatusCode != check.status {
			return fmt.Errorf("reverse driver did not reject wrong authentication: %d", rsp.StatusCode)
		}
	}
	conn, _, err := websocket.DefaultDialer.Dial(target, http.Header{"Authorization": []string{"Bearer " + token}})
	if err != nil {
		return fmt.Errorf("reverse driver authenticated upgrade failed: %w", err)
	}
	defer conn.Close()
	if err = conn.WriteJSON(map[string]any{"self_id": 87654323, "post_type": "meta_event", "meta_event_type": "lifecycle", "sub_type": "connect", "time": time.Now().Unix()}); err != nil {
		return err
	}
	code, err := os.ReadFile(filepath.Join(dataDir, "setup-code.txt"))
	if err != nil {
		return err
	}
	setup, _ := json.Marshal(map[string]string{"setupCode": strings.TrimSpace(string(code)), "password": "reverse-fixture-password-only"})
	req, _ := http.NewRequest("POST", base+"/api/setup", bytes.NewReader(setup))
	req.Header.Set("Origin", base)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Local-Request", "1")
	rsp, err := client.Do(req)
	if err != nil {
		return err
	}
	io.Copy(io.Discard, rsp.Body)
	rsp.Body.Close()
	if rsp.StatusCode != 200 {
		return fmt.Errorf("reverse fixture setup failed: %d", rsp.StatusCode)
	}
	connected := false
	for attempt := 0; attempt < 50; attempt++ {
		rsp, e := client.Get(base + "/api/overview")
		if e == nil {
			var overview map[string]any
			decodeErr := json.NewDecoder(rsp.Body).Decode(&overview)
			rsp.Body.Close()
			accounts, _ := overview["botAccounts"].([]any)
			if decodeErr == nil && rsp.StatusCode == 200 && overview["connectionState"] == "connected" && len(accounts) == 1 && accounts[0] == "87654323" {
				connected = true
				break
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	if !connected {
		return fmt.Errorf("reverse handshake did not register synthetic account")
	}
	rsp, err = client.Get(base + "/api/connections")
	if err != nil {
		return err
	}
	var settings map[string]any
	err = json.NewDecoder(rsp.Body).Decode(&settings)
	rsp.Body.Close()
	if err != nil || rsp.StatusCode != 200 {
		return fmt.Errorf("reverse fixture settings unavailable")
	}
	views, _ := settings["runningConnections"].([]any)
	if len(views) != 1 || views[0].(map[string]any)["kind"] != "wss" || views[0].(map[string]any)["tokenConfigured"] != true {
		return fmt.Errorf("reverse driver startup config mismatch")
	}
	encoded, _ := json.Marshal(settings)
	if bytes.Contains(encoded, []byte(token)) {
		return fmt.Errorf("reverse fixture read exposed token")
	}
	if err = conn.WriteJSON(map[string]any{"time": time.Now().Unix(), "self_id": 87654323, "post_type": "message", "message_type": "private", "sub_type": "friend", "message_id": 1, "user_id": 12345678, "raw_message": "/help", "message": []map[string]any{{"type": "text", "data": map[string]string{"text": "/help"}}}, "sender": map[string]any{"user_id": 12345678, "nickname": "Reverse synthetic user", "sex": "unknown", "age": 0}}); err != nil {
		return err
	}
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	help := false
	for i := 0; i < 30; i++ {
		var request apiRequest
		if err = conn.ReadJSON(&request); err != nil {
			return fmt.Errorf("reverse command reply missing: %w", err)
		}
		if err = conn.WriteJSON(map[string]any{"status": "ok", "retcode": 0, "data": map[string]any{"message_id": 999, "user_id": 87654323, "nickname": "Reverse synthetic bot"}, "echo": request.Echo}); err != nil {
			return err
		}
		if request.Action == "send_private_msg" {
			text, e := smokeMessageText(request.Params.Message)
			if e != nil || request.Params.UserID != 12345678 || !strings.Contains(text, "OneBot + ZeroBot + Golang") {
				return fmt.Errorf("reverse command delivery incorrect")
			}
			help = true
			break
		}
	}
	if !help {
		return fmt.Errorf("reverse command response not observed")
	}
	evidence := map[string]any{"passed": true, "isolated": true, "realQQMessagesSent": false, "missingTokenDenied": true, "wrongTokenDenied": true, "authenticatedUpgradePassed": true, "actualReverseAccountRegistered": true, "startupConfigAndRedactionPassed": true, "helpRoundtripPassed": true, "completedAt": time.Now().Format(time.RFC3339)}
	encoded, _ = json.MarshalIndent(evidence, "", "  ")
	return os.WriteFile(filepath.Join(fixture, "result.json"), encoded, 0600)
}
