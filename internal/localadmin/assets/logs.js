"use strict";
let logActive=false, logPaused=false, logCursor="0", logInstance="", logGeneration=0;
let logTimer=null, logRequest=null, logLevel="", logKeyword="", logSeen=new Set();
let logExportRequest=null;
let logFrom="", logUntil="";
const logCopyEntries=new WeakMap();
const logLevelLabels={"":"全部级别",issues:"警告与错误",panic:"Panic",fatal:"Fatal",error:"错误",warning:"警告",info:"信息",debug:"调试",trace:"跟踪"};

function renderLogFilters() {
  const draft=$("log-level").value!==logLevel || $("log-keyword").value.trim()!==logKeyword || $("log-from").value!==logFrom || $("log-until").value!==logUntil;
  const range=logFrom || logUntil ? "时间（本地，含边界）："+(logFrom.replace("T"," ") || "不限")+" 至 "+(logUntil.replace("T"," ") || "不限")+"。" : "未限制时间。";
  $("log-filter-status").textContent="已应用："+logLevelLabels[logLevel]+"；"+(logKeyword ? "关键词“"+logKeyword+"”。" : "未限制关键词。")+range+(draft ? " 输入修改尚未应用；查看与导出仍使用上述条件。" : "");
}
function logTimeValue(value) {
  if (!value) return "";
  const date=new Date(value);
  // Reject normalization of impossible calendar or daylight-saving local times.
  const match=/^(\d{4})-(\d{2})-(\d{2})T(\d{2}):(\d{2})(?::(\d{2}))?$/.exec(value);
  if (!match || !Number.isFinite(date.getTime()) || Number(match[1])<1 || [date.getFullYear(),date.getMonth()+1,date.getDate(),date.getHours(),date.getMinutes(),date.getSeconds()].some((n,i)=>n!==Number(match[i+1] || 0))) throw new Error("日志时间无效，请填写本地有效日期与时间。");
  return date.toISOString();
}
function appliedLogQuery() {
  const query=new URLSearchParams({level:logLevel,q:logKeyword});
  if (logFrom) query.set("from",logTimeValue(logFrom));
  if (logUntil) query.set("until",logTimeValue(logUntil));
  return query;
}
function logLocalTime(timestamp) {
  const date=new Date(timestamp), pad=n=>String(n).padStart(2,"0");
  const value=String(date.getFullYear()).padStart(4,"0")+"-"+pad(date.getMonth()+1)+"-"+pad(date.getDate())+"T"+pad(date.getHours())+":"+pad(date.getMinutes())+":"+pad(date.getSeconds());
  // datetime-local cannot distinguish the second occurrence of a repeated DST hour.
  if (Date.parse(logTimeValue(value))!==timestamp) throw new Error("当前快捷范围包含无法明确表示的本地时间，请手动选择起止时间。");
  return value;
}
function applyRecentLogRange(minutes) {
  if (!authenticated || !logActive || document.hidden || ![15,60,1440].includes(minutes)) return;
  try {
    const end=Math.floor(Date.now()/1000)*1000;
    const from=logLocalTime(end-minutes*60000),until=logLocalTime(end);
    applyLogFilters(logLevel,logKeyword,from,until);
  } catch(error) {showError("log-filter-error",error);renderLogFilters();}
}
function applyLogFilters(level,keyword,from="",until="") {
  try {
    const start=logTimeValue(from),end=logTimeValue(until);
    if (start && end && start>end) throw new Error("开始时间不能晚于结束时间。");
  } catch(error) {showError("log-filter-error",error);renderLogFilters();return;}
  $("log-filter-error").hidden=true;
  logLevel=level;logKeyword=keyword.trim();logFrom=from;logUntil=until;
  $("log-level").value=logLevel;$("log-keyword").value=logKeyword;$("log-from").value=logFrom;$("log-until").value=logUntil;
  renderLogFilters();reloadLogs();
}

function stopLogRequests() {
  logGeneration++;
  clearTimeout(logTimer); logTimer=null;
  if (logRequest) logRequest.abort();
  logRequest=null;
  if (logExportRequest) logExportRequest.abort();
  logExportRequest=null;
  $("export-logs").disabled=false; $("log-export-status").textContent="";
  renderLogReadOnce();
  renderLogCopyControls(true);
}
function resetLogs() {
  stopLogRequests(); logActive=false; logPaused=false; logCursor="0"; logInstance=""; logLevel=""; logKeyword=""; logSeen.clear();
  logFrom="";logUntil="";$("log-from").value="";$("log-until").value="";$("log-filter-error").hidden=true;
  $("log-records").replaceChildren(); $("log-keyword").value=""; $("log-level").value="";
  $("pause-logs").textContent="暂停更新"; $("log-gap").hidden=true; $("log-error").hidden=true;
  $("log-status").textContent="进入页面后读取日志。";
  $("log-clear-status").textContent="";
  clearLogSummary();
  renderLogFilters();
  renderLogReadOnce();
}
function showLogs(active) {
  stopLogRequests(); logActive=active;
  renderLogReadOnce();
  renderLogCopyControls();
  if (active && authenticated && !document.hidden) {
    if (!logPaused) pollLogs();
    else $("log-status").textContent="已暂停更新；恢复后读取保留范围内的新记录。";
  }
}
function renderLogs(value) {
  renderLogSummary(value);
  const container=$("log-records");
  if (value.reset) {container.replaceChildren(); logSeen.clear(); $("log-gap").hidden=false;$("log-clear-status").textContent="";}
  if (value.gap) $("log-gap").hidden=false;
  const empty=container.querySelector(".empty"); if (empty) empty.remove();
  for (const entry of value.entries) {
    if (logSeen.has(entry.id)) continue;
    logSeen.add(entry.id);
    const row=document.createElement("article"), meta=document.createElement("div"), time=document.createElement("time"), level=document.createElement("span"), text=document.createElement("pre");
    row.className="log-entry"; row.dataset.id=entry.id; meta.className="log-meta";
    time.dateTime=entry.time; time.textContent=new Date(entry.time).toLocaleString("zh-CN",{hour12:false});
    level.className="log-level"; level.textContent=entry.level.toUpperCase();
    if (["error","fatal","panic"].includes(entry.level)) level.classList.add("log-error-level");
    if (entry.level==="warning") level.classList.add("log-warning-level");
    text.textContent=entry.message;text.tabIndex=-1;
    const actions=document.createElement("div"),copy=document.createElement("button"),status=document.createElement("span");
    actions.className="log-copy-actions";copy.type="button";copy.className="secondary copy-log-entry";copy.textContent="复制本条";
    copy.setAttribute("aria-label","复制 "+time.textContent+" 的 "+level.textContent+" 日志");
    status.className="field-help";status.setAttribute("role","status");status.setAttribute("aria-live","polite");
    logCopyEntries.set(copy,{row,time,level,text,status,request:null});copy.addEventListener("click",()=>copyLogEntry(copy));
    actions.append(copy,status);meta.append(time,level);row.append(meta,text,actions);container.append(row);
    while (container.children.length>1000) {const old=container.firstElementChild; logSeen.delete(old.dataset.id); old.remove();}
  }
  if (!container.children.length) {const empty=document.createElement("p"); empty.className="empty"; empty.textContent="暂时没有符合筛选条件的日志。"; container.append(empty);}
  logCursor=value.cursor; logInstance=value.instance;
  $("log-status").textContent=(logPaused ? "已暂停" : "持续更新")+" · 当前显示 "+logSeen.size+" 条 / 服务保留 "+value.retained+" 条";
  renderLogCopyControls();
  renderLogClearControls();
  if ($("log-autoscroll").checked) container.scrollTop=container.scrollHeight;
}
function renderLogClearControls() {
  $("clear-log-display").disabled=!authenticated || !logActive || document.hidden || !!logRequest || !logSeen.size;
}
function clearLogDisplay() {
  if (!authenticated || !logActive || document.hidden || logRequest || !logSeen.size) return;
  // Keep the instance and cursor: this clears only records already displayed.
  renderLogCopyControls(true);
  logSeen.clear();
  const empty=document.createElement("p");empty.className="empty";empty.textContent="已清空当前显示；后续沿原游标读取，服务保留记录未删除。";
  $("log-records").replaceChildren(empty);
  $("log-clear-status").textContent="仅清空本页显示，匹配统计和导出仍涵盖服务保留记录；重新加载后可读取近期日志。";
  $("log-status").textContent=(logPaused ? "已暂停" : "持续更新")+" · 当前显示 0 条（仅清空页面）。";
  renderLogClearControls();
}
function logCopyReady(record) {
  return authenticated && logActive && !document.hidden && record.row.isConnected && $("log-records").contains(record.row);
}
function renderLogCopyControls(reset=false) {
  if (reset) {
    const selection=window.getSelection?.();
    if (selection?.anchorNode && $("log-records").contains(selection.anchorNode)) selection.removeAllRanges();
  }
  for (const button of $("log-records").querySelectorAll(".copy-log-entry")) {
    const record=logCopyEntries.get(button);if (!record) continue;
    if (reset) {record.request=null;record.status.textContent="";}
    button.disabled=!logCopyReady(record) || !!record.request;
  }
}
function logCopyCurrent(record,snapshot) {
  return record.request===snapshot && snapshot.generation===logGeneration && snapshot.token===csrf && logCopyReady(record);
}
async function copyLogEntry(button) {
  const record=logCopyEntries.get(button);
  if (!record || record.request || !logCopyReady(record)) return;
  const snapshot={generation:logGeneration,token:csrf};record.request=snapshot;button.disabled=true;
  // Copy exactly the displayed text, never the raw API entry or extra fields.
  const content=record.time.textContent+" ["+record.level.textContent+"]\n"+record.text.textContent;
  record.status.textContent="正在复制…";
  try {
    if (!navigator.clipboard?.writeText) throw new Error("clipboard unavailable");
    await navigator.clipboard.writeText(content);
    if (logCopyCurrent(record,snapshot)) record.status.textContent="已复制本条日志。";
  } catch {
    if (logCopyCurrent(record,snapshot)) {
      const range=document.createRange();range.selectNodeContents(record.text);
      const selection=window.getSelection();selection?.removeAllRanges();selection?.addRange(range);
      record.text.focus({preventScroll:true});
      record.status.textContent="复制未完成，已选中日志正文；请按 Ctrl+C 手动复制。";
    }
  } finally {
    if (logCopyCurrent(record,snapshot)) {record.request=null;button.disabled=false;}
  }
}
function clearLogSummary() {
  $("log-summary-state").textContent = "读取当前筛选后显示统计。";
  $("log-summary-levels").replaceChildren();
}
function renderLogSummary(value) {
  const summary = value.summary, levels = summary?.levels;
  const names = [["panic","Panic"],["fatal","Fatal"],["error","错误"],["warning","警告"],["info","信息"],["debug","调试"],["trace","跟踪"],["other","其他"]];
  const count = n => Number.isInteger(n) && n >= 0 && n <= 1000;
  const valid = summary && count(value.retained) && count(summary.matched) && summary.matched <= value.retained && levels && names.every(([name]) => count(levels[name])) && names.reduce((sum,[name]) => sum + levels[name],0) === summary.matched;
  const list = $("log-summary-levels"); list.replaceChildren();
  if (!valid) { $("log-summary-state").textContent = "当前接口未提供可核对的匹配统计，不按页面条数推断。"; return; }
  $("log-summary-state").textContent = "符合已应用筛选：" + summary.matched + " / 服务仍保留：" + value.retained + " 条（最近成功读取于 " + new Date().toLocaleTimeString("zh-CN",{hour12:false}) + "）。";
  for (const [name,label] of names) {
    const item = document.createElement("span"); item.textContent = label + "：" + levels[name]; list.append(item);
  }
}
function renderLogReadOnce() {
  $("read-logs-once").disabled=!logActive || !authenticated || !logPaused || document.hidden || !!logRequest;
  renderLogClearControls();
}
async function readLogsOnce() {
  if (!logPaused) return;
  await pollLogs(true);
}
async function pollLogs(once=false) {
  if (!logActive || (logPaused && once!==true) || !authenticated || document.hidden || logRequest) return;
  const generation=logGeneration, controller=new AbortController(); logRequest=controller;
  renderLogReadOnce();
  if (once===true) $("log-status").textContent="正在读取一次；自动更新仍保持暂停。";
  const query=appliedLogQuery();query.set("after",logCursor);query.set("instance",logInstance);query.set("limit","200");
  try {
    const value=await api("/api/logs?"+query,undefined,controller.signal);
    if (generation!==logGeneration || !authenticated || !logActive) return;
    $("log-error").hidden=true; renderLogs(value);
  } catch (error) {
    if (error.name!=="AbortError" && generation===logGeneration && authenticated && logActive) {
      showError("log-error",error); $("log-status").textContent=logPaused ? "读取未完成；仍保持暂停，可再次点击读取一次。" : "读取中断，稍后自动重试。";
    }
  } finally {
    if (logRequest===controller) logRequest=null;
    if (generation===logGeneration) renderLogReadOnce();
    if (generation===logGeneration && logActive && !logPaused && authenticated && !document.hidden) logTimer=setTimeout(pollLogs,2000);
  }
}
function reloadLogs() {
  stopLogRequests(); logCursor="0"; logInstance=""; logSeen.clear(); $("log-records").replaceChildren(); $("log-gap").hidden=true;
  clearLogSummary();
  $("log-clear-status").textContent="";
  renderLogClearControls();
  if (!logPaused) pollLogs(); else $("log-status").textContent="已暂停更新；恢复后应用当前筛选。";
}
async function exportLogs() {
  if (!logActive || !authenticated || document.hidden || logExportRequest) return;
  const generation=logGeneration, controller=new AbortController(); logExportRequest=controller;
  $("export-logs").disabled=true; $("log-export-status").textContent="正在准备脱敏日志…";
  try {
    const query=appliedLogQuery();
    const response=await fetch("/api/logs/export?"+query,{credentials:"same-origin",signal:controller.signal});
    if (generation!==logGeneration || !authenticated || !logActive || document.hidden) return;
    if (!response.ok) {
      const value=await response.json();
      if (generation!==logGeneration || !authenticated || !logActive) return;
      if (response.status===401) {authenticated=false; csrf=""; showAuth(); return;}
      throw new Error(value.error || "日志导出失败");
    }
    const disposition=response.headers.get("Content-Disposition") || "";
    const filename=/^attachment; filename="(zerobot-logs-\d{8}T\d{6}Z\.txt)"$/.exec(disposition)?.[1];
    const maxBytes=3*1024*1024;
    const length=response.headers.get("Content-Length") || "";
    if (!filename || !response.headers.get("Content-Type")?.startsWith("text/plain") || !/^\d+$/.test(length) || Number(length)>maxBytes) throw new Error("日志导出格式异常");
    const content=await response.blob();
    if (generation!==logGeneration || !authenticated || !logActive || document.hidden) return;
    if (content.size>maxBytes) throw new Error("日志导出超过容量限制");
    const url=URL.createObjectURL(content), link=document.createElement("a");
    try {
      link.href=url; link.download=filename; document.body.append(link); link.click();
      $("log-export-status").textContent="已发起下载；文件仅包含仍保留的脱敏日志。";
    } finally {link.remove(); setTimeout(()=>URL.revokeObjectURL(url),0);}
  } catch (error) {
    if (error.name!=="AbortError" && generation===logGeneration && authenticated && logActive) $("log-export-status").textContent=error.message;
  } finally {
    if (logExportRequest===controller) {logExportRequest=null; $("export-logs").disabled=false;}
  }
}
$("export-logs").addEventListener("click",exportLogs);
$("read-logs-once").addEventListener("click",readLogsOnce);
$("clear-log-display").addEventListener("click",clearLogDisplay);
for (const [id,minutes] of [["log-recent-15m",15],["log-recent-1h",60],["log-recent-24h",1440]]) $(id).addEventListener("click",()=>applyRecentLogRange(minutes));
$("log-any-time").addEventListener("click",()=>{if (authenticated && logActive && !document.hidden) applyLogFilters(logLevel,logKeyword);});
$("log-filter").addEventListener("submit",event=>{event.preventDefault(); if (authenticated && logActive) applyLogFilters($("log-level").value,$("log-keyword").value,$("log-from").value,$("log-until").value);});
$("only-issue-logs").addEventListener("click",()=>{if (authenticated && logActive) applyLogFilters("issues",logKeyword,logFrom,logUntil);});
$("clear-log-filter").addEventListener("click",()=>{if (authenticated && logActive) applyLogFilters("","");});
$("view-issue-logs").addEventListener("click",()=>{
  if (!authenticated) return;
  // Set conditions before navigation so the first visible request uses them.
  stopLogRequests();logActive=false;applyLogFilters("issues","");
  document.querySelector('.nav-button[data-view="logs"]').click();
});
$("log-level").addEventListener("change",renderLogFilters);
$("log-keyword").addEventListener("input",renderLogFilters);
for (const id of ["log-from","log-until"]) $(id).addEventListener("input",()=>{$("log-filter-error").hidden=true;renderLogFilters();});
$("reload-logs").addEventListener("click",reloadLogs);
$("pause-logs").addEventListener("click",()=>{
  logPaused=!logPaused; stopLogRequests(); $("pause-logs").textContent=logPaused ? "恢复更新" : "暂停更新";
  if (logPaused) $("log-status").textContent="已暂停更新；恢复后读取保留范围内的新记录。"; else pollLogs();
});
$("log-autoscroll").addEventListener("change",()=>{if ($("log-autoscroll").checked) $("log-records").scrollTop=$("log-records").scrollHeight;});
document.addEventListener("visibilitychange",()=>{stopLogRequests(); if (!document.hidden) pollLogs();});
window.addEventListener("pagehide",stopLogRequests);
