"use strict";
let connectionAdding = false, connectionAddGeneration = 0;
let connectionAddConfirmation = null;
function finishConnectionAddConfirmation(accepted = false) {
  const pending = connectionAddConfirmation;
  connectionAddConfirmation = null;
  const dialog = $("connection-add-confirmation");
  if (dialog.open) dialog.close();
  $("connection-add-confirm-message").textContent = "";
  if (!pending) return;
  pending.resolve(accepted);
  if (authenticated && pending.origin?.isConnected && !pending.origin.disabled) pending.origin.focus();
}
function confirmConnectionAddition(message, label) {
  if (connectionAddConfirmation || connectionDraftConfirmation || !authenticated || connectionBusy) return Promise.resolve(false);
  return new Promise(resolve => {
    connectionAddConfirmation = {resolve, origin:document.activeElement};
    $("connection-add-confirm-message").textContent = message;
    $("connection-add-confirm-accept").textContent = label;
    try { $("connection-add-confirmation").showModal(); }
    catch (_) {
      finishConnectionAddConfirmation(false);
      showError("connection-add-error", new Error("无法打开确认窗口，请重新加载页面后再操作。"));
    }
  });
}
$("connection-add-confirm-cancel").addEventListener("click", () => finishConnectionAddConfirmation(false));
$("connection-add-confirm-accept").addEventListener("click", () => finishConnectionAddConfirmation(true));
$("connection-add-confirmation").addEventListener("cancel", event => {
  event.preventDefault(); finishConnectionAddConfirmation(false);
});
$("connection-add-confirmation").addEventListener("close", () => {
  if (!$("connection-add-confirmation").open) finishConnectionAddConfirmation(false);
});
function connectionAddSnapshot() {
  return JSON.stringify([$("connection-add-kind").value,$("connection-add-url").value, $("connection-add-token").value, $("connection-add-without-token").checked]);
}
function connectionAddDirty() {
  return !!connectionSettings?.available && !!($("connection-add-kind").value!=="ws" || $("connection-add-url").value || $("connection-add-token").value || $("connection-add-without-token").checked);
}
function renderConnectionAddControls() {
  const blocked = !authenticated || connectionBusy || !connectionSettings?.available;
  $("connection-add-fields").disabled = blocked;
  const reverse=$("connection-add-kind").value==="wss";
  $("connection-add-without-token").disabled=reverse;
  $("connection-add-kind-note").textContent=reverse ? "仅支持 ws://127.0.0.1:明确端口和根路径，必须填写 Token；重启后在接入程序配置相同地址与 Token。" : "正向连接支持 ws/wss 接入地址，可明确选择不使用 Token。";
  $("connection-add-token").disabled = $("connection-add-without-token").checked;
  $("connection-add-token").required = !$("connection-add-without-token").checked;
  $("connection-add-save").disabled = blocked || !connectionAddDirty();
  $("connection-add-discard").disabled = blocked || !connectionAddDirty();
}
function resetConnectionAdd() {
  finishConnectionAddConfirmation(false);
  connectionAddGeneration++; connectionAdding = false;
  $("connection-add-url").value = ""; $("connection-add-token").value = "";
  $("connection-add-kind").value="ws"; $("connection-add-without-token").checked = false; $("connection-add-error").hidden = true;
  $("connection-add-status").textContent = "新增与现有连接/主人分别保存；重启后生效。";
  renderConnectionAddControls();
}
function connectionAddCanLeave() {
  if (connectionAddConfirmation || connectionDraftConfirmation) return false;
  if (!connectionAdding) return true;
  $("connection-add-status").textContent = "新增请求尚未完成，请稍候再切换页面。";
  return false;
}
$("connection-add-kind").addEventListener("change",()=>{
  if($("connection-add-kind").value==="wss") $("connection-add-without-token").checked=false;
  renderConnectionDrafts();
});
$("connection-add-without-token").addEventListener("change", () => {
  if ($("connection-add-without-token").checked) $("connection-add-token").value = "";
  renderConnectionDrafts();
});
for (const event of ["input", "change"]) $("connection-add-form").addEventListener(event, () => {
  $("connection-add-status").textContent = "新增草稿未保存；保存不会自动连接或重启。";
  renderConnectionDrafts();
});
$("connection-add-discard").addEventListener("click", async () => {
  if (!authenticated || connectionBusy || !connectionAddDirty()) return;
  const generation=connectionGeneration, addGeneration=connectionAddGeneration, draft=connectionAddSnapshot(), token=csrf, settings=connectionSettings;
  if (!await confirmConnectionAddition("放弃新增连接的未保存更改？已有连接、主人草稿与已保存配置保持不变。", "确认放弃")) return;
  if (!authenticated || connectionBusy || token!==csrf || settings!==connectionSettings || generation!==connectionGeneration || addGeneration!==connectionAddGeneration || draft!==connectionAddSnapshot()) return;
  resetConnectionAdd(); renderConnectionDrafts(); $("connection-add-status").textContent="已放弃新增草稿，已有配置保持不变。";
});
$("connection-add-form").addEventListener("submit", async event => {
  event.preventDefault();
  if (!authenticated || connectionBusy || !connectionSettings?.available) return;
  $("connection-add-error").hidden = true;
  if (connectionDirty(true) || connectionOwnersDirty()) { showError("connection-add-error",new Error("请先保存或放弃现有连接与主人账号的未保存更改，再新增连接。")); return; }
  const settings=connectionSettings, generation=connectionGeneration, addGeneration=connectionAddGeneration, token=csrf, draft=connectionAddSnapshot();
  const kind=$("connection-add-kind").value;
  const addConnection={url:$("connection-add-url").value,token:$("connection-add-token").value,withoutToken:$("connection-add-without-token").checked};
  if(kind!=="ws" && kind!=="wss") {showError("connection-add-error",new Error("请选择有效的新增连接方向。"));return;}
  if(kind==="wss") addConnection.kind="wss";
  if(kind==="wss" && (addConnection.withoutToken || !addConnection.token)) {showError("connection-add-error",new Error("本机反向监听必须填写 Token。"));return;}
  if (!addConnection.url || (!addConnection.withoutToken && !addConnection.token)) { showError("connection-add-error",new Error("填写新连接地址及 Token，或明确选择不使用 Token。")); return; }
  if (!await confirmConnectionAddition("新增"+(kind==="wss" ? "本机反向监听" : "正向连接")+"并备份当前配置？保存不自动连接或监听，接入程序请使用相同地址与 Token。", "确认新增")) return;
  if (!authenticated || connectionBusy || token!==csrf || settings!==connectionSettings || generation!==connectionGeneration || addGeneration!==connectionAddGeneration || draft!==connectionAddSnapshot()) return;
  connectionAdding = true; connectionControls(true);
  $("connection-add-status").textContent = "正在备份并保存新连接…";
  try {
    const value = await api("/api/connections",{revision:settings.revision,addConnection});
    if (!authenticated || token!==csrf || generation!==connectionGeneration || addGeneration!==connectionAddGeneration) return;
    const index=(value.connections || []).findIndex(item=>item.kind===kind && item.index===(settings.connections || []).filter(item=>item.kind===kind).length);
    if (!value.available || !value.backupId || !value.revision || index<0 || value.connections.length!==(settings.connections || []).length+1) throw new Error("新增结果无法确认，请重新读取配置核对后再操作。");
    connectionIndex=index; resetConnectionAdd(); renderConnectionSettings(value);
    $("connection-save-state").textContent += " 旧配置备份已创建。";
    $("connection-add-status").textContent = value.restartRequired ? "新连接已保存并选中，待重启机器人生效；新增 Token 输入已清空。" : "新连接已保存，配置与本次启动一致，无需为本次新增重启；Token 输入已清空。";
  } catch (error) {
    if (authenticated && token===csrf && generation===connectionGeneration && addGeneration===connectionAddGeneration) {
      showError("connection-add-error",error);
      $("connection-add-status").textContent="新增未确认成功，草稿保留；若网络中断，请先重新读取配置核对。";
    }
  } finally {
    // Successful save resets the addition generation; shared connection generation
    // still identifies the same authenticated session and its busy ownership.
    if (authenticated && token===csrf && generation===connectionGeneration) {
      connectionAdding=false; connectionControls(false); renderConnectionDrafts();
    }
  }
});
