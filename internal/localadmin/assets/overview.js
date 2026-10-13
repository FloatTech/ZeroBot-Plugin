"use strict";
let overviewActive=false, overviewBusy=false, overviewGeneration=0, overviewRequest=null, overviewTimer=null;

function clearModelRequests(message="请求统计待更新。") {
  if (typeof clearModelObservationCurrent==="function") clearModelObservationCurrent();
  $("model-request-summary").textContent=message;
  $("model-request-categories").replaceChildren();
}
function renderModelRequests(value) {
  const names={gpt:"GPT / 群总结",chat:"普通聊天",agent:"Agent",vision:"识图"};
  const counts=["started","active","completed","failed","incomplete"];
  const valid=row=>row && counts.every(key=>Number.isSafeInteger(row[key]) && row[key]>=0) &&
    row.started===row.active+row.completed+row.failed+row.incomplete &&
    ["averageMilliseconds","maxMilliseconds"].every(key=>Number.isFinite(row[key]) && row[key]>=0) && row.maxMilliseconds>=row.averageMilliseconds;
  if (!valid(value) || !Array.isArray(value.categories) || value.categories.length!==4 ||
      !Object.keys(names).every(kind=>value.categories.filter(row=>row?.kind===kind && valid(row)).length===1) ||
      !counts.every(key=>value[key]===value.categories.reduce((sum,row)=>sum+row[key],0)) ||
      typeof value.startedAt!=="string" || !Number.isFinite(Date.parse(value.startedAt)) ||
      typeof value.checkedAt!=="string" || !Number.isFinite(Date.parse(value.checkedAt))) {
    clearModelRequests("请求统计暂不可读取。");return;
  }
  const elapsed=row=>row.completed+row.failed+row.incomplete ?
    "已结束请求平均 "+(row.averageMilliseconds/1000).toFixed(3)+" 秒，最长 "+(row.maxMilliseconds/1000).toFixed(3)+" 秒。" : "尚无已结束请求，耗时待统计。";
  const countsText=row=>"发起 "+row.started+" 次；进行中 "+row.active+" 次；接口完成 "+row.completed+" 次；失败 "+row.failed+" 次；提前关闭 "+row.incomplete+" 次。";
  $("model-request-summary").textContent="统计起点："+new Date(value.startedAt).toLocaleString("zh-CN",{hour12:false})+"。"+countsText(value)+" "+elapsed(value)+(value.started===0 ? " 本次启动尚无模型接口请求。" : "");
  $("model-request-categories").replaceChildren();
  for (const [kind,name] of Object.entries(names)) {
    const row=value.categories.find(item=>item.kind===kind),card=document.createElement("article"),title=document.createElement("h3"),count=document.createElement("p"),time=document.createElement("p");
    title.textContent=name;count.textContent=countsText(row);time.className="field-help";time.textContent=elapsed(row);card.append(title,count,time);$("model-request-categories").append(card);
  }
  if (typeof receiveModelObservation==="function") receiveModelObservation(value);
}

function renderOverviewAlerts(value) {
  $("overview-alerts").replaceChildren();
  if (!value) {
    $("overview-alert-summary").textContent="近期异常暂不可读取，请到运行日志检查配置是否可读。";
    return;
  }
  $("overview-alert-summary").textContent="本次启动仍保留的日志：警告 "+value.warningCount+" 条，错误 "+value.errorCount+" 条。"+(value.evicted ? " 更早记录已淘汰。" : "")+" 显示最新 5 条；历史记录不代表问题仍未解决。";
  if (!value.entries.length) {
    const empty=document.createElement("p");empty.className="empty";empty.textContent="保留日志中暂无警告或错误。";$("overview-alerts").append(empty);
  }
  for (const entry of value.entries) {
    const row=document.createElement("article"),meta=document.createElement("div"),time=document.createElement("time"),level=document.createElement("span"),text=document.createElement("pre");
    row.className="log-entry";meta.className="log-meta";
    time.textContent=new Date(entry.time).toLocaleString("zh-CN",{hour12:false});
    level.textContent=entry.level.toUpperCase();level.className="log-level "+(entry.level==="warning" ? "log-warning-level" : "log-error-level");
    text.textContent=entry.message;meta.append(time,level);row.append(meta,text);$("overview-alerts").append(row);
  }
}
function overviewReady(generation) {return authenticated && overviewActive && !document.hidden && generation===overviewGeneration;}
function staleOverview(message="状态待更新") {
  if (typeof clearReplyTasks==="function") clearReplyTasks();
  clearModelRequests();
  $("connection-badge").textContent=message;$("connection-badge").classList.add("waiting");
  $("connection-detail").textContent="请等待重新读取当前连接状态。";
  $("account-count").textContent="—";$("checked-at").textContent="—";$("accounts").replaceChildren();
  $("overview-alerts").replaceChildren();$("overview-alert-summary").textContent="近期异常待更新。";
}
function cancelOverview() {
  overviewGeneration++;clearTimeout(overviewTimer);overviewTimer=null;
  overviewRequest?.abort();overviewRequest=null;overviewBusy=false;$("refresh").disabled=false;
  if (typeof renderModelObservation==="function") renderModelObservation();
  if (typeof renderReplyDownload==="function") renderReplyDownload();
}
function resetOverview() {
  overviewActive=false;cancelOverview();staleOverview();$("page-error").hidden=true;
  if (typeof resetModelObservation==="function") resetModelObservation();
}
function showOverview(active) {
  overviewActive=active;cancelOverview();staleOverview();
  $("page-error").hidden=true;
  if (active && authenticated && !document.hidden) refreshOverview();
}
async function refreshOverview() {
  if (!authenticated || !overviewActive || document.hidden || overviewBusy) return;
  clearTimeout(overviewTimer);overviewTimer=null;
  const generation=++overviewGeneration,controller=new AbortController();overviewRequest=controller;overviewBusy=true;$("refresh").disabled=true;$("page-error").hidden=true;
  if (typeof renderModelObservation==="function") renderModelObservation();
  if (typeof renderReplyDownload==="function") renderReplyDownload();
  try {
    const value=await api("/api/overview",undefined,controller.signal);
    if (overviewReady(generation)) renderOverview(value);
  } catch(error) {
    if (overviewReady(generation) && error.name!=="AbortError") {staleOverview("无法获取状态");$("connection-detail").textContent="状态更新失败，旧连接信息已清除。";showError("page-error",error);}
  } finally {
    if (overviewReady(generation)) {overviewRequest=null;overviewBusy=false;$("refresh").disabled=false;overviewTimer=setTimeout(refreshOverview,10000);}
    if (overviewReady(generation) && typeof renderModelObservation==="function") renderModelObservation();
    if (overviewReady(generation) && typeof renderReplyDownload==="function") renderReplyDownload();
  }
}
document.addEventListener("visibilitychange",()=>{
  cancelOverview();staleOverview();
  if (!document.hidden && authenticated && overviewActive) refreshOverview();
});
