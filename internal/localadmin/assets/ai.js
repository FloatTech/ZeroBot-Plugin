"use strict";
let aiState = null, aiBusy = false, aiGroup = "0", aiPrivate = "", aiGeneration = 0;
const aiFields = [
  ["Type","聊天接口类型","type"], ["ModelName","聊天模型名","text"], ["API","聊天接口地址","text"],
  ["MaxN","最大输出长度（0 为默认 4096）","number",0,131072], ["TopP","TopP（0 为默认 0.9）","number",0,1],
  ["SystemP","系统提示词","textarea"], ["NoSystemP","不发送系统提示词","checkbox"],
  ["ReasoningEffort","推理努力度（DeepSeek 官方接口：none 关闭思考，low/high/max 开启；留空用模型默认）","text"], ["Separator","回复分隔符（可留空）","text"],
  ["ImageType","识图接口类型","type"], ["ImageModelName","识图模型名（可留空）","text"], ["ImageAPI","识图接口地址（可留空）","text"],
  ["AgentType","Agent 接口类型","type"], ["AgentModelName","Agent 模型名（可留空）","text"], ["AgentAPI","Agent 接口地址（可留空）","text"],
  ["AgentChar","Agent 人物设定（修改后保留对话记录）","textarea"], ["AgentSex","Agent 性别","text"]
];
const aiKeys = [["Key","聊天密钥"],["ImageKey","识图密钥"],["AgentKey","Agent 密钥"]];
const aiBaselines = new Map();
function aiFormSnapshot(form) {
  return JSON.stringify([...form.querySelectorAll("input,select,textarea")].map(input=>[input.id,input.type==="checkbox" ? input.checked : input.value]));
}
function aiDrafts() {
  return [...aiBaselines.entries()].filter(([,entry])=>entry.form.isConnected && aiFormSnapshot(entry.form)!==entry.value).map(([name])=>name);
}
function renderAIDrafts() {
  const drafts=aiDrafts();
  renderAIDraftFields(drafts);
  $("ai-draft-status").textContent=drafts.length ? "未保存更改："+drafts.map(name=>({global:"全局模型与人物设定",aichatcfg:"聊天会话参数",llm:"GPT / 群总结温度"}[name])).join("、")+"。请分别保存；切换后台页面会保留草稿。" : aiState ? "当前 AI 表单没有未保存的更改。" : "读取设置后可编辑；草稿仅保留在当前页面。";
  for (const button of document.querySelectorAll("[data-ai-discard]")) button.disabled = !authenticated || aiBusy || !drafts.includes(button.dataset.aiDiscard);
}
function discardAIForm(name) {
  if (!authenticated || aiBusy || !aiDrafts().includes(name)) return;
  const entry = aiBaselines.get(name), generation = aiGeneration;
  const title = entry.form.querySelector("h2")?.textContent || "此表单";
  if (!window.confirm("放弃“" + title + "”的未保存更改？其他表单的编辑和已保存配置保持不变。")) return;
  if (!authenticated || aiBusy || generation !== aiGeneration || aiBaselines.get(name) !== entry || !entry.form.isConnected) return;
  const values = new Map(JSON.parse(entry.value));
  for (const input of entry.form.querySelectorAll("input,select,textarea")) {
    if (!values.has(input.id)) continue;
    if (input.type === "checkbox") input.checked = values.get(input.id); else input.value = values.get(input.id);
  }
  if (name === "global") {
    // Reapply the key action UI: empty replacement input, keep mode, no clear confirmation.
    for (const [key] of aiKeys) $("ai-action-" + key).dispatchEvent(new Event("change", {bubbles:true}));
  }
  $("ai-status").textContent = "已放弃“" + title + "”的未保存更改；已保存配置保持不变。";
  renderAIDrafts();
}
function aiChangedFieldNames(name,entry) {
  const baseline=new Map(JSON.parse(entry.value)),current=new Map(JSON.parse(aiFormSnapshot(entry.form)));
  const changed=new Set([...new Set([...baseline.keys(),...current.keys()])].filter(id=>baseline.get(id)!==current.get(id)));
  const labels=[];
  const take=(ids,label)=>{if(ids.some(id=>changed.has(id)))labels.push(label);for(const id of ids)changed.delete(id);};
  if (name==="global") {
    for (const [field,label] of aiFields) take(["ai-field-"+field],label);
    for (const [key,label] of aiKeys) take(["ai-action-"+key,"ai-key-"+key,"ai-clear-"+key],label);
  } else if (name==="aichatcfg" || name==="llm") {
    for (const [field,label] of [["temperature","温度百分比"],...(name==="aichatcfg" ? [["rate","非 @ 消息触发概率"],["agent","Agent 模式"],["voice","QQ AI 语音输出"],["at","响应 @ 消息"]] : [])]) take(["ai-field-"+field+"-"+name],label);
  }
  // Only fixed labels reach the DOM: never display input IDs or old/new values.
  if (changed.size) labels.push("其他表单字段");
  return labels;
}
function aiDraftFieldScope(name,entry) {
  if (name==="global") return "全局";
  const scope=entry.scope,valid=id=>typeof id==="string" && /^[1-9][0-9]{0,18}$/.test(id);
  if (valid(scope?.privateId) && (!scope.groupId || scope.groupId==="0")) return "私聊 QQ "+scope.privateId;
  if (!scope?.privateId && valid(scope?.groupId)) return "群 "+scope.groupId;
  return "范围暂不可核对";
}
function renderAIDraftFields(drafts) {
  const panel=$("ai-draft-fields"),list=$("ai-draft-field-list");list.replaceChildren();
  panel.hidden=!authenticated || !aiState?.available || !drafts.length;
  $("ai-draft-field-state").textContent="";
  if (panel.hidden) return;
  $("ai-draft-field-state").textContent=aiBusy ? "请求处理中，保留当前字段摘要，请等待结果；不会自动保存。" : "与该表单最近读取或保存成功的页面基线比较；只列字段名称，不表示已生效或服务端最新状态。";
  for (const name of drafts) {
    const entry=aiBaselines.get(name),labels=aiChangedFieldNames(name,entry);
    const item=document.createElement("article"),heading=document.createElement("h3"),count=document.createElement("p"),fields=document.createElement("ul");
    const scope=aiDraftFieldScope(name,entry),title=name==="llm" && scope.startsWith("私聊 QQ ") ? "GPT 温度" : ({global:"全局模型与人物设定",aichatcfg:"聊天会话参数",llm:"GPT / 群总结温度"}[name] || "其他表单");
    item.dataset.aiDraftForm=["global","aichatcfg","llm"].includes(name) ? name : "other";heading.textContent=title+" · "+scope;
    count.className="field-help";count.textContent="未保存字段："+labels.length+" 项（密钥操作、输入与确认合并计数）。";
    for (const label of labels) {const field=document.createElement("li");field.textContent=label;fields.append(field);}
    item.append(heading,count,fields);list.append(item);
  }
}
function rememberAIForm(name,form,scope=null) {
  aiBaselines.set(name,{form,value:aiFormSnapshot(form),scope}); renderAIDrafts();
}
function confirmAIDiscard(action) {
  if (aiBusy) { window.alert("AI 设置请求尚未结束，请等待结果后再"+action+"。"); return false; }
  return !aiDrafts().length || window.confirm("AI 设置有未保存的更改。"+action+"将放弃这些更改，是否继续？");
}

function resetAI() {
  aiGeneration++; aiState = null; aiGroup = "0"; aiPrivate = "";
  aiBaselines.clear(); aiBusy=false;
  renderAIConfigChecks(null);
  $("ai-global-inputs").replaceChildren(); $("ai-group-forms").replaceChildren(); $("ai-group").value = "";
  $("ai-scope-kind").value = "group"; renderAIScopeLabel();
  aiBusyState(false); renderAIDrafts();
}
function aiBusyState(busy) {
  aiBusy = busy; $("reload-ai").disabled = busy; $("apply-ai-group").disabled = busy;
  $("ai-global-fields").disabled = busy || !aiState?.available;
  for (const fields of $("ai-group-forms").querySelectorAll("fieldset")) fields.disabled = busy;
  renderAIDrafts();
}
function aiInput(container, name, labelText, kind, value, min, max) {
  const label = document.createElement("label"), input = document.createElement(kind === "textarea" ? "textarea" : kind === "type" ? "select" : "input");
  input.id = "ai-field-" + name; label.htmlFor = input.id; label.textContent = labelText;
  if (kind === "type") {
    for (const [index, text] of ["OpenAI", "OLLaMA", "GenAI"].entries()) { const option = document.createElement("option"); option.value = String(index); option.textContent = text; input.append(option); }
    if (![0,1,2].includes(value)) { const option = document.createElement("option"); option.value = String(value); option.textContent = "原配置类型无效，请选择"; input.append(option); }
    input.value = String(value);
  } else if (kind === "checkbox") { input.type = "checkbox"; input.checked = value; label.className = "check-label"; label.prepend(input); container.append(label); return input; }
  else {
    if (kind !== "textarea") input.type = kind;
    input.value = value ?? ""; input.autocomplete = "off";
    if (kind === "number") { input.min = min; input.max = max; input.step = name === "TopP" ? "0.01" : "1"; input.required = true; }
    else { input.maxLength = kind === "textarea" ? 32768 : name.endsWith("API") || name === "API" ? 2048 : 256; if (kind === "textarea") input.rows = 6; }
  }
  container.append(label, input); return input;
}
function renderAIGlobal(value) {
  renderAIConfigChecks(value);
  if (!value.available) { $("ai-status").textContent = "当前程序没有编译 AI 配置插件。"; $("ai-global-inputs").replaceChildren(); aiBaselines.delete("global"); renderAIDrafts(); return; }
  const container = $("ai-global-inputs"); container.replaceChildren();
  for (const [name,label,kind,min,max] of aiFields) {
    aiInput(container,name,label,kind,value.global.fields[name],min,max);
    if (value.global.addressRedacted[name]) { const note = document.createElement("p"); note.className="field-help"; note.textContent="旧地址中的认证参数已隐藏；不修改时保留原地址，替换时请填写独立密钥。"; container.append(note); }
  }
  for (const [name,text] of aiKeys) {
    const label=document.createElement("label"), select=document.createElement("select"), key=document.createElement("input"), note=document.createElement("p"), confirm=document.createElement("label"), checkbox=document.createElement("input");
    select.id="ai-action-"+name; label.htmlFor=select.id; label.textContent=text;
    for (const [action,title] of [["keep","保留当前密钥"],["replace","替换为新密钥"],["clear","清除密钥"]]) { const option=document.createElement("option"); option.value=action; option.textContent=title; select.append(option); }
    key.id="ai-key-"+name; key.type="password"; key.autocomplete="off"; key.maxLength=4096; key.placeholder="输入新密钥"; key.setAttribute("aria-label",text+"新值"); key.disabled=true;
    note.className="field-help"; note.textContent=value.global.keys[name] ? "已配置；当前值不会显示。" : "未配置。";
    confirm.className="check-label"; confirm.hidden=true; checkbox.type="checkbox"; checkbox.id="ai-clear-"+name; confirm.append(checkbox,document.createTextNode("我确认清除"+text));
    select.addEventListener("change",()=>{key.disabled=select.value!=="replace"; key.required=select.value==="replace"; if (select.value!=="replace") key.value=""; confirm.hidden=select.value!=="clear"; checkbox.checked=false;});
    container.append(label,select,note,key,confirm);
  }
  $("ai-status").textContent=value.global.ready ? "配置已读取。保存后的设置供后续请求使用。" : "聊天接口尚未配置完整；OLLaMA 可使用无密钥接口。";
  rememberAIForm("global",$("ai-global-form"));
}
function renderAIConfigChecks(value) {
  const panel = $("ai-config-checks"), list = $("ai-config-check-list");
  list.replaceChildren(); panel.hidden = !authenticated || !value?.available;
  if (panel.hidden) return;
  for (const [name,title] of [["chat","聊天 / GPT"],["agent","Agent"],["image","识图（可选）"]]) {
    const item = document.createElement("article"), heading = document.createElement("h3"), status = document.createElement("p"), detail = document.createElement("p");
    item.dataset.aiCheck = name; heading.textContent = title; status.className = "ai-check-result"; detail.className = "field-help";
    const check = value.global?.checks?.[name];
    const fields = ["typeValid","modelConfigured","modelRequired","addressConfigured","addressValid","keyConfigured","keyRequired","complete"];
    const valid = check && fields.every(key => typeof check[key] === "boolean") && check.modelRequired === (name !== "image") &&
      (!check.addressValid || check.addressConfigured) && check.complete === (check.typeValid && check.addressConfigured && check.addressValid && (check.modelConfigured || !check.modelRequired) && (check.keyConfigured || !check.keyRequired));
    if (!valid) {
      status.textContent = "暂不可核对基础状态"; detail.textContent = "接口未提供可核对的检查结果，请重新加载或更新程序。";
    } else {
      const optionalEmpty = name === "image" && check.typeValid && !check.modelConfigured && !check.addressConfigured && !check.keyConfigured;
      status.textContent = optionalEmpty ? "未配置（可选，不阻止文字聊天）" : check.complete ? "基础字段检查通过；尚未验证实际调用" : "基础字段待补充或检查";
      detail.textContent = "接口类型：" + (check.typeValid ? "支持" : "需检查") + "；模型：" + (check.modelConfigured ? "已填写" : check.modelRequired ? "未填写（需要）" : "未填写（可选，按接口要求核对）") +
        "；地址：" + (!check.addressConfigured ? "未填写" : check.addressValid ? "格式检查通过" : "需检查格式或隐藏认证参数") + "；密钥：" + (check.keyConfigured ? "已配置（未验证）" : check.keyRequired ? "未配置（需要）" : "未配置（OLLaMA 可无密钥）") + "。";
    }
    item.append(heading, status, detail); list.append(item);
  }
}
function renderAIGroups(groups, onlyService="") {
  const container=$("ai-group-forms"); if (!onlyService) {container.replaceChildren(); aiBaselines.delete("aichatcfg"); aiBaselines.delete("llm");}
  for (const value of groups) {
    if (onlyService && value.service !== onlyService) continue;
    const form=document.createElement("form"), fields=document.createElement("fieldset"), heading=document.createElement("h2"), save=document.createElement("button"); form.className="panel ai-form";
    form.dataset.service=value.service;
    heading.textContent=(value.service==="llm" ? (value.privateId ? "GPT 温度" : "GPT / 群总结温度") : "聊天会话参数")+" · "+(value.privateId ? "私聊 QQ "+value.privateId : "群 "+value.groupId); fields.append(heading);
    const suffix=value.service;
    const temp=aiInput(fields,"temperature-"+suffix,"温度百分比（0 为默认 70）","number",value.temperature,0,100);
    let rate,agent,voice,at;
    if (value.service==="aichatcfg") {
      rate=aiInput(fields,"rate-"+suffix,"非 @ 消息触发概率（百分比）","number",value.rate,0,100);
      agent=aiInput(fields,"agent-"+suffix,"使用 Agent 模式（可调用机器人操作）","checkbox",value.useAgent);
      voice=aiInput(fields,"voice-"+suffix,"尝试使用 QQ AI 语音输出","checkbox",value.voiceOutput);
      at=aiInput(fields,"at-"+suffix,"响应 @ 消息","checkbox",value.respondAt);
    }
    save.type="submit"; save.className="primary"; save.textContent="备份并保存会话参数";
    const discard = document.createElement("button"); discard.type="button"; discard.className="secondary ai-discard"; discard.textContent="放弃此表单更改";
    discard.id="discard-ai-"+value.service; discard.dataset.aiDiscard=value.service; discard.addEventListener("click",()=>discardAIForm(value.service));
    fields.append(save,discard); form.append(fields);
    const existing=onlyService ? [...container.children].find(item=>item.dataset.service===value.service) : null;
    if (existing) existing.replaceWith(form); else container.append(form);
    rememberAIForm(value.service,form,{groupId:value.groupId,privateId:value.privateId});
    form.addEventListener("submit",async(event)=>{
      event.preventDefault(); if (aiBusy || !authenticated) return;
      const generation=aiGeneration, body={service:value.service,revision:value.revision,temperature:Number(temp.value)};
      if (value.privateId) body.privateId=value.privateId; else body.groupId=value.groupId;
      if (rate) Object.assign(body,{rate:Number(rate.value),useAgent:agent.checked,voiceOutput:voice.checked,respondAt:at.checked});
      aiBusyState(true); $("ai-error").hidden=true;
      try { const saved=await api("/api/ai-group",body); if (generation!==aiGeneration || !authenticated) return; aiState.groups=saved.groups; renderAIGroups(saved.groups,value.service); $("ai-status").textContent="会话参数已保存；后续消息使用新参数。"; }
      catch(error) { if (generation===aiGeneration && authenticated) showError("ai-error",error); }
      finally { if (generation===aiGeneration) {aiBusyState(false); renderAIDrafts();} }
    });
  }
}
async function loadAI(group=aiGroup,privateId=aiPrivate) {
  if (aiBusy || !authenticated) return;
  if (!confirmAIDiscard(group===aiGroup && privateId===aiPrivate ? "重新加载设置" : "切换会话参数")) return;
  const generation=++aiGeneration; aiBusyState(true); $("ai-error").hidden=true;
  try { const value=await api("/api/ai?"+(privateId ? "private="+encodeURIComponent(privateId) : "group="+encodeURIComponent(group))); if (generation!==aiGeneration || !authenticated) return; aiGroup=group; aiPrivate=privateId; aiState=value; $("ai-scope-kind").value=privateId ? "private" : "group"; if (privateId || group!=="0") $("ai-group").value=privateId || group; renderAIScopeLabel(); renderAIGlobal(value); renderAIGroups(value.groups); }
  catch(error) { if (generation===aiGeneration && authenticated) { $("ai-status").textContent="读取失败，请重新加载。"; showError("ai-error",error); } }
  finally { if (generation===aiGeneration) {aiBusyState(false); renderAIDrafts();} }
}
$("reload-ai").addEventListener("click",()=>loadAI());
$("discard-ai-global").addEventListener("click",()=>discardAIForm("global"));
function renderAIScopeLabel() {
  const privateScope=$("ai-scope-kind").value==="private";
  $("ai-conversation-label").textContent=privateScope ? "私聊用户 QQ" : "群号";
  $("ai-group").placeholder=privateScope ? "填写与机器人聊天的用户 QQ" : "填写要管理的群号";
}
$("ai-scope-kind").addEventListener("change",renderAIScopeLabel);
$("ai-scope-form").addEventListener("submit",(event)=>{ event.preventDefault(); const id=$("ai-group").value.trim(); if (!/^[1-9][0-9]{0,18}$/.test(id)) {showError("ai-error",new Error("群号或私聊 QQ 须为正整数。")); return;} loadAI($("ai-scope-kind").value==="private" ? "0" : id,$("ai-scope-kind").value==="private" ? id : ""); });
$("ai-global-form").addEventListener("submit",async(event)=>{
  event.preventDefault(); if (aiBusy || !aiState?.available || !authenticated) return;
  const generation=aiGeneration, body={revision:aiState.global.revision,fields:{},keys:{}};
  for (const [name,,kind] of aiFields) {
    const input=$("ai-field-"+name), value=kind==="checkbox" ? input.checked : kind==="number" || kind==="type" ? Number(input.value) : input.value;
    if (value !== aiState.global.fields[name]) body.fields[name]=value;
  }
  for (const [name] of aiKeys) {
    const action=$("ai-action-"+name).value; body.keys[name]={action};
    if (action==="replace") body.keys[name].value=$("ai-key-"+name).value;
    if (action==="clear") { if (!$("ai-clear-"+name).checked) {showError("ai-error",new Error("请先确认清除密钥。")); return;} body.keys[name].confirm=true; }
  }
  aiBusyState(true); $("ai-error").hidden=true;
  try { const saved=await api("/api/ai",body); if (generation!==aiGeneration || !authenticated) return; aiState.global=saved.global; renderAIGlobal(aiState); $("ai-status").textContent="全局配置已保存，后续请求使用新设置，无需重启。"; }
  catch(error) { if (generation===aiGeneration && authenticated) showError("ai-error",error); }
  finally { if (generation===aiGeneration) {aiBusyState(false); renderAIDrafts();} }
});
for (const id of ["ai-global-form","ai-group-forms"]) {
  $(id).addEventListener("input",renderAIDrafts); $(id).addEventListener("change",renderAIDrafts);
}
window.addEventListener("beforeunload",event=>{
  if (authenticated && (aiBusy || aiDrafts().length)) {event.preventDefault();event.returnValue="";}
});
