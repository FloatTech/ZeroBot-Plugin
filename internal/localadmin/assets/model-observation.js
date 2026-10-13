"use strict";
let modelObservationCurrent=null, modelObservationBaseline=null, modelObservationDownloading=false, modelObservationGeneration=0;
const modelObservationKinds={gpt:"GPT / 群总结",chat:"普通聊天",agent:"Agent",vision:"识图"};
const modelObservationCounts=["started","active","completed","failed","incomplete"];
const modelObservationTimes=["averageMilliseconds","maxMilliseconds"];

function modelObservationStamp(value) {
  if (typeof value!=="string") return null;
  const match=value.match(/^(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2})(?:\.(\d{1,9}))?Z$/);
  if (!match || !Number.isFinite(Date.parse(value)) || new Date(value).toISOString().slice(0,19)!==match[1]) return null;
  return match[1]+"."+(match[2]||"").padEnd(9,"0")+"Z";
}
function copyModelObservation(value) {
  const copy=row=>{
    if (!row || !modelObservationCounts.every(key=>Number.isSafeInteger(row[key]) && row[key]>=0) ||
        row.started!==row.active+row.completed+row.failed+row.incomplete ||
        !modelObservationTimes.every(key=>Number.isFinite(row[key]) && row[key]>=0) || row.averageMilliseconds>row.maxMilliseconds ||
        (row.completed+row.failed+row.incomplete===0 && (row.averageMilliseconds!==0 || row.maxMilliseconds!==0))) return null;
    return Object.fromEntries([...modelObservationCounts,...modelObservationTimes].map(key=>[key,row[key]]));
  };
  const totals=copy(value),startedAt=modelObservationStamp(value?.startedAt),checkedAt=modelObservationStamp(value?.checkedAt);
  if (!totals || !startedAt || !checkedAt || checkedAt<startedAt || !Array.isArray(value.categories) || value.categories.length!==4) return null;
  const categories=[];
  for (const kind of Object.keys(modelObservationKinds)) {
    const rows=value.categories.filter(row=>row?.kind===kind);
    if (rows.length!==1) return null;
    const counts=copy(rows[0]);if (!counts) return null;
    categories.push({kind,...counts});
  }
  if (!modelObservationCounts.every(key=>totals[key]===categories.reduce((sum,row)=>sum+row[key],0))) return null;
  let process;
  if (value.process!=null) {
    const source=value.process,processStartedAt=modelObservationStamp(source.processStartedAt);
    if (!Number.isSafeInteger(source.processId) || source.processId<=0 || !processStartedAt || processStartedAt>startedAt ||
        typeof source.programSHA256!=="string" || !/^[a-fA-F0-9]{64}$/.test(source.programSHA256)) return null;
    process={processId:source.processId,processStartedAt,programSHA256:source.programSHA256.toUpperCase()};
  }
  return {startedAt,checkedAt,...totals,categories,...(process ? {process} : {})};
}
function modelObservationDelta(before,after) {
  if (!before || !after || before.startedAt!==after.startedAt || after.checkedAt<before.checkedAt) return null;
  if (JSON.stringify(before.process||null)!==JSON.stringify(after.process||null)) return null;
  if (after.checkedAt===before.checkedAt && JSON.stringify(before)!==JSON.stringify(after)) return null;
  const compare=(start,end)=>{
    const delta={};
    for (const key of ["started","completed","failed","incomplete"]) {
      if (end[key]<start[key]) return null;
      delta[key]=end[key]-start[key];
    }
    if (end.maxMilliseconds<start.maxMilliseconds || delta.started+start.active!==delta.completed+delta.failed+delta.incomplete+end.active) return null;
    return {...delta,activeAtStart:start.active,activeAtEnd:end.active};
  };
  const totals=compare(before,after);if (!totals) return null;
  const categories=[];
  for (let i=0;i<4;i++) {
    const row=compare(before.categories[i],after.categories[i]);if (!row) return null;
    categories.push({kind:before.categories[i].kind,...row});
  }
  return {durationSeconds:(Date.parse(after.checkedAt)-Date.parse(before.checkedAt))/1000,...totals,categories};
}
function modelObservationReady() { return authenticated && overviewActive && !document.hidden && !overviewBusy && !modelObservationDownloading && !!modelObservationCurrent; }
function renderModelObservation() {
  const ready=modelObservationReady(),delta=modelObservationDelta(modelObservationBaseline,modelObservationCurrent);
  $("start-model-observation").disabled=!ready;
  $("start-model-observation").textContent=modelObservationBaseline ? "重新记录观察起点" : "记录观察起点";
  $("clear-model-observation").disabled=!authenticated || !modelObservationBaseline || modelObservationDownloading;
  $("download-model-observation").disabled=!ready || !delta;
  if (!modelObservationBaseline) { $("model-observation-summary").textContent="尚未记录观察起点。";return; }
  if (!modelObservationCurrent) { $("model-observation-summary").textContent="观察起点已记录，结束快照待重新读取；当前不可下载。";return; }
  if (!delta) { $("model-observation-summary").textContent="前后统计无法核对，请重新记录观察起点。";return; }
  $("model-observation-summary").textContent="观察快照间隔 "+delta.durationSeconds.toFixed(3)+" 秒：新发起 "+delta.started+" 次；结束并完成接口传输 "+delta.completed+" 次；失败 "+delta.failed+" 次；提前关闭 "+delta.incomplete+" 次。起点进行中 "+delta.activeAtStart+" 次，结束进行中 "+delta.activeAtEnd+" 次。";
}
function clearModelObservationCurrent() { modelObservationCurrent=null;renderModelObservation(); }
function receiveModelObservation(value) {
  modelObservationCurrent=copyModelObservation(value);
  if (modelObservationBaseline && modelObservationCurrent && !modelObservationDelta(modelObservationBaseline,modelObservationCurrent)) {
    modelObservationBaseline=null;
    $("model-observation-status").textContent="机器人进程、统计起点、读取时间或累计计数已改变，原观察起点已清除，请重新记录。";
  }
  renderModelObservation();
}
function resetModelObservation() {
  modelObservationGeneration++;modelObservationCurrent=null;modelObservationBaseline=null;modelObservationDownloading=false;
  $("model-observation-status").textContent="";renderModelObservation();
}
function startModelObservation() {
  if (!modelObservationReady()) return;
  modelObservationBaseline=copyModelObservation(modelObservationCurrent);
  $("model-observation-status").textContent="已记录最近成功读取的统计快照；自动刷新后可下载前后报告。"+(modelObservationBaseline.process ? "报告包含进程关联信息。" : "进程关联信息不可读取，报告仅含次数与时间。")+"仅保存在当前页面，刷新页面或退出后清除。";
  renderModelObservation();
}
function clearModelObservation() {
  if (!authenticated || !modelObservationBaseline || modelObservationDownloading) return;
  modelObservationBaseline=null;$("model-observation-status").textContent="观察起点已清除，机器人累计统计保持。";renderModelObservation();
}
function downloadModelObservation() {
  if (!modelObservationReady()) return;
  const before=copyModelObservation(modelObservationBaseline),after=copyModelObservation(modelObservationCurrent),delta=modelObservationDelta(before,after);
  if (!delta) return;
  const exportedAt=new Date().toISOString(),report={format:"zerobot-model-observation",version:before.process ? 2 : 1,exportedAt,
    scope:{kind:"model-http-exchanges",latencyScope:"process-cumulative",businessAcceptancePassed:false},
    before,after,window:delta};
  let url=null,link=null;
  const generation=modelObservationGeneration,token=csrf;
  modelObservationDownloading=true;renderModelObservation();
  try {
    url=URL.createObjectURL(new Blob([JSON.stringify(report,null,2)+"\n"],{type:"application/json;charset=utf-8"}));
    link=document.createElement("a");link.href=url;link.download="zerobot-model-observation-"+exportedAt.replace(/[-:.]/g,"")+".json";document.body.append(link);link.click();
    $("model-observation-status").textContent="已发起观察报告下载；包括前后累计快照与请求次数变化。累计平均/最长耗时仍属于本次启动，不是窗口最长耗时，也不证明QQ回复成功。";
  } catch {
    $("model-observation-status").textContent="下载未完成，请稍后重试。";
  } finally {
    link?.remove();
    setTimeout(()=>{
      if (url) URL.revokeObjectURL(url);
      if (generation===modelObservationGeneration && authenticated && csrf===token) {modelObservationDownloading=false;renderModelObservation();}
    },1000);
  }
}
$("start-model-observation").addEventListener("click",startModelObservation);
$("clear-model-observation").addEventListener("click",clearModelObservation);
$("download-model-observation").addEventListener("click",downloadModelObservation);
