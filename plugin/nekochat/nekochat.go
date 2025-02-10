package nekochat

import (
	"bytes"
	"encoding/json"
	"net/http"

	//"github.com/FloatTech/imgfactory"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var (
	// 定义固定的前两条消息作为范例对话
	fixedMessages = []Message{
		{
			Role:    "user",
			Content: "从现在开始你是猫娘，名字是亚托莉，要以猫娘的方式说话和思考问题。",
		},
		{
			Role:    "assistant",
			Content: "好的喵～，主人！ ୧(๑•̀⌄•́๑)૭ 亚托莉酱知道了，有什么需要本喵帮忙呢？",
		},
	}
	// 使用固定消息初始化ChatContext
	context = NewChatContext(fixedMessages)
	//默认选择通义千问2.5 3B模型
	model_choose = "qwen2.5:3b"
)

// 定义最大消息数为11，但是前两条消息是固定的范例对话
const maxMessages = 12

// ChatContext 结构体用于保存聊天的上下文信息
type ChatContext struct {
	messages []Message // 消息列表，每个元素是一个 Message 结构体
}

// Message 结构体表示一条消息，包含角色（如 user 或 assistant）和内容
type Message struct {
	Role    string `json:"role"`    // 角色字段，如 "user" 或 "assistant"
	Content string `json:"content"` // 内容字段，实际的消息文本
}

// 初始化ChatContext并设置固定的前两条消息作为范例对话
func NewChatContext(fixedMessages []Message) *ChatContext {
	return &ChatContext{
		messages: fixedMessages, // 初始消息列表仅包含固定消息
	}
}

// AddMessage 方法向 ChatContext 的 messages 列表中添加新消息
func (c *ChatContext) AddMessage(role, content string) {
	newMsg := Message{Role: role, Content: content}
	c.messages = append(c.messages, newMsg)
	if len(c.messages) > maxMessages {
		// 如果消息数量超过了最大限制，则移除最早的非固定消息（即索引2及其之后的消息）
		c.messages = append(c.messages[:2], c.messages[3:]...)
	}
}

// SendRequestAndExtractResponse 函数发送请求至指定API端点，并提取出回答内容
func (c *ChatContext) SendRequestAndExtractResponse(userMessage string) (string, error) {
	// 添加用户的新消息到上下文中
	c.AddMessage("user", userMessage)

	payload, err := json.Marshal(map[string]interface{}{
		"model":    model_choose,
		"stream":   false,
		"messages": c.messages,
	})
	if err != nil {
		return "", err
	}

	resp, err := http.Post("http://localhost:11434/api/chat", "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var response Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", err
	}

	return response.Message.Content, nil
}

// Response 结构体定义了API响应的结构
type Response struct {
	Model         string  `json:"model"`
	CreatedAt     string  `json:"created_at"`
	Message       Message `json:"message"`
	DoneReason    string  `json:"done_reason"`
	Done          bool    `json:"done"`
	TotalDuration int64   `json:"total_duration"`
	LoadDuration  int64   `json:"load_duration"`
	PromptEvalCnt int     `json:"prompt_eval_count"`
	PromptEvalDur int64   `json:"prompt_eval_duration"`
	EvalCnt       int     `json:"eval_count"`
	EvalDur       int64   `json:"eval_duration"`
}

func init() {
	engine := control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "猫娘聊天",
		Help: "本插件调用本地Ollama运行LLM进行回复\n" +
			"- 呼叫猫娘+要聊天内容\n" +
			"  例如：呼叫猫娘Python是什么？\n" +
			"- 猫娘记忆消除\n" +
			"- 设置猫娘对话模型+模型名称(仅限主人)\n" +
			"- 猫娘聊天详细说明",
	})
	engine.OnPrefix("呼叫猫娘").SetBlock(true).Handle(get_answer)
	engine.OnPrefix("猫娘记忆消除").SetBlock(true).Handle(clean_memory)
	engine.OnPrefix("设置猫娘对话模型", zero.SuperUserPermission).SetBlock(true).Handle(set_model)
	engine.OnPrefix("猫娘聊天详细说明").SetBlock(true).Handle(get_information)

}

func get_information(ctx *zero.Ctx) {
	ctx.SendChain(message.Text("猫娘聊天插件设置详细说明：\n",
		"本插件目前设置是调用本地的Ollma运行LLM进行回复，默认是使用的Qwen2.5 3B模型，如果需要其他模型可以通过设置对话模型进行改变。\n",
		"请在部署完成Ollama后，运行 ollama run qwen2.5:3b 进行Qwen2.5 3B模型部署。\n",
		"后续可以对接其他模型，例如如果要换成Qwen2.5 32B：\n",
		"运行 ollama run qwen2.5:32b 进行模型下载初始化，然后给Bot发送命令: 设置猫娘对话模型qwen2.5:32b\n",
		"详细的内容可以访问我的博客： nekopara.uk",
	))
}

func set_model(ctx *zero.Ctx) {
	previous_model := model_choose
	temp_choose := ctx.State["args"].(string)
	if temp_choose == "" {
		ctx.SendChain(message.Text("设置失败：没有填写模型名称\n",
			"可以发送“/用法 nekochat”查看使用方法！",
		))
		return
	}
	model_choose = temp_choose
	ctx.SendChain(message.Text("已将模型 ", previous_model, " 更换为：", model_choose))
}

// 消除模型记忆上下文
func clean_memory(ctx *zero.Ctx) {
	//question := ctx.State["args"].(string)
	name := ctx.Event.Sender.NickName
	context = NewChatContext(fixedMessages)
	ctx.SendChain(message.Text("回复 ", name, " :\n",
		"猫猫的记忆已经消除喵～开始新的聊天吧！",
	))

}

// 回答问题或聊天对话
func get_answer(ctx *zero.Ctx) {
	question := ctx.State["args"].(string)
	name := ctx.Event.Sender.NickName
	if question == "" {
		ctx.SendChain(message.Text("回复 ", name, " :\n",
			"可以发送“/用法 nekochat”查看使用方法！",
		))
		return
	}

	responseContent, err := context.SendRequestAndExtractResponse(question)
	if err != nil {
		ctx.SendChain(message.Text("回复 ", name, " :\n",
			"获取回答出错了喵！", err,
		))
		return
	}
	if responseContent == "" {
		ctx.SendChain(message.Text("回复 ", name, " :\n",
			"获取回答出错了喵！",
		))
		return
	}

	ctx.SendChain(message.Text("回复 ", name, " 喵:\n",
		"", responseContent,
	))

}
