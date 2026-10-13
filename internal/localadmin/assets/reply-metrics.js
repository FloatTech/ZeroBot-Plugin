"use strict";
const replyTaskCounts=["started","active","completed","failed","noReply","withReply","missingIngress","replyAttempts","replyAccepted","replyUnconfirmed"];
const replyTaskTimes=["averageFirstReplyMilliseconds","maxFirstReplyMilliseconds","averageLastReplyMilliseconds","maxLastReplyMilliseconds"];
let replyTaskCurrent=null,replyDownloadBusy=false,replyDownloadGeneration=0;
function clearReplyTasks(message="回复处理统计待更新。") {
  replyTaskCurrent=null;replyDownloadGeneration++;replyDownloadBusy=false;
  $("reply-task-summary").textContent=message;$("reply-task-categories").replaceChildren();
  $("reply-download-status").textContent="";renderReplyDownload();
}
function replyTaskStamp(value) {
  if (typeof value!=="string") return null;
  const m=value.match(/^(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2})(?:\.(\d{1,9}))?Z$/);
  if (!m || !Number.isFinite(Date.parse(value)) || new Date(value).toISOString().slice(0,19)!==m[1]) return null;
  return m[1]+"."+(m[2]||"").padEnd(9,"0")+"Z";
}
function copyReplyTasks(value) {
  const kinds=["gpt","chat","agent"];
  const valid=row=>row && replyTaskCounts.every(k=>Number.isSafeInteger(row[k])&&row[k]>=0) &&
    BigInt(row.started)===BigInt(row.active)+BigInt(row.completed)+BigInt(row.failed)+BigInt(row.noReply) &&
    BigInt(row.replyAttempts)===BigInt(row.replyAccepted)+BigInt(row.replyUnconfirmed) && row.missingIngress<=row.started &&
    BigInt(row.withReply)<=BigInt(row.completed)+BigInt(row.failed) && row.completed<=row.withReply && row.withReply<=row.replyAccepted &&
    replyTaskTimes.every(k=>Number.isFinite(row[k])&&row[k]>=0) && row.averageFirstReplyMilliseconds<=row.maxFirstReplyMilliseconds &&
    row.averageLastReplyMilliseconds<=row.maxLastReplyMilliseconds && row.averageLastReplyMilliseconds>=row.averageFirstReplyMilliseconds && row.maxLastReplyMilliseconds>=row.maxFirstReplyMilliseconds &&
    (row.withReply!==0||replyTaskTimes.every(k=>row[k]===0));
  const start=replyTaskStamp(value?.startedAt),checked=replyTaskStamp(value?.checkedAt);
  if (!valid(value)||value.scope!=="bot-matcher-entry-to-onebot-send-receipt"||!start||!checked||checked<start||
      !Array.isArray(value.categories)||value.categories.length!==3||!kinds.every(kind=>value.categories.filter(row=>row?.kind===kind&&valid(row)).length===1)||
      !replyTaskCounts.every(k=>BigInt(value[k])===value.categories.reduce((n,row)=>n+BigInt(row[k]),0n))) return null;
  let process;
  if (value.process!=null) {
    const p=value.process,processStartedAt=replyTaskStamp(p.processStartedAt);
    if (!Number.isSafeInteger(p.processId)||p.processId<=0||!processStartedAt||processStartedAt>start||typeof p.programSHA256!=="string"||!/^[a-fA-F0-9]{64}$/.test(p.programSHA256)) return null;
    process={processId:p.processId,processStartedAt,programSHA256:p.programSHA256.toUpperCase()};
  }
  const copy=row=>Object.fromEntries([...replyTaskCounts,...replyTaskTimes].map(k=>[k,row[k]]));
  return {startedAt:start,checkedAt:checked,scope:value.scope,...copy(value),categories:kinds.map(kind=>({kind,...copy(value.categories.find(row=>row.kind===kind))})),...(process?{process}:{})};
}
function renderReplyTasks(value) {
  value=copyReplyTasks(value);
  if (!value) { clearReplyTasks("回复处理统计暂不可读取。");return; }
  replyTaskCurrent=value;renderReplyDownload();
  const names={gpt:"GPT / 群总结",chat:"普通聊天",agent:"Agent / 识图"};
  const counts=row=>"任务 "+row.started+" 次；进行中 "+row.active+" 次；取得回复回执 "+row.completed+" 次；失败 "+row.failed+" 次；没有回复 "+row.noReply+" 次。";
  const timings=row=>row.withReply ? "已有确认回复的已结束任务 "+row.withReply+" 次：首条平均 "+(row.averageFirstReplyMilliseconds/1000).toFixed(3)+" 秒，最长 "+(row.maxFirstReplyMilliseconds/1000).toFixed(3)+" 秒；最后一条平均 "+(row.averageLastReplyMilliseconds/1000).toFixed(3)+" 秒，最长 "+(row.maxLastReplyMilliseconds/1000).toFixed(3)+" 秒。" : "尚无已结束且取得确认回复的任务，耗时待统计。";
  const receipts=row=>"当前会话回复发送 "+row.replyAttempts+" 次；确认 "+row.replyAccepted+" 次；未确认 "+row.replyUnconfirmed+" 次。缺少匹配起点 "+row.missingIngress+" 次。";
  $("reply-task-summary").textContent="统计起点："+new Date(value.startedAt).toLocaleString("zh-CN",{hour12:false})+"。"+counts(value)+" "+timings(value)+" "+receipts(value);
  $("reply-task-categories").replaceChildren();
  for (const [kind,name] of Object.entries(names)) {
    const row=value.categories.find(item=>item.kind===kind),card=document.createElement("article"),title=document.createElement("h3"),count=document.createElement("p"),time=document.createElement("p"),receipt=document.createElement("p");
    title.textContent=name;count.textContent=counts(row);time.className=receipt.className="field-help";time.textContent=timings(row);receipt.textContent=receipts(row);card.append(title,count,time,receipt);$("reply-task-categories").append(card);
  }
}
function renderReplyDownload() {
  $("download-reply-statistics").disabled=!authenticated||!overviewActive||document.hidden||overviewBusy||replyDownloadBusy||!replyTaskCurrent;
}
function downloadReplyStatistics() {
  renderReplyDownload();
  if ($("download-reply-statistics").disabled) return;
  const snapshot=copyReplyTasks(replyTaskCurrent);
  if (!snapshot) return;
  const exportedAt=new Date().toISOString(),report={format:"zerobot-reply-statistics",version:1,exportedAt,
    scope:{kind:"ai-reply-tasks",latencyScope:"process-cumulative",timingBoundary:"bot-matcher-entry-to-onebot-send-receipt",businessAcceptancePassed:false},snapshot};
  const generation=replyDownloadGeneration,token=csrf;let url=null,link=null;
  replyDownloadBusy=true;renderReplyDownload();
  try {
    url=URL.createObjectURL(new Blob([JSON.stringify(report,null,2)+"\n"],{type:"application/json;charset=utf-8"}));
    link=document.createElement("a");link.href=url;link.download="zerobot-reply-statistics-"+exportedAt.replace(/[-:.]/g,"")+".json";document.body.append(link);link.click();
    $("reply-download-status").textContent="已发起回复统计下载，内容为最近读取的累计快照；请核对文件中的读取时间。平均/最长值属于本次启动，不是观察窗口耗时，也不证明QQ客户端收到。";
  } catch {
    $("reply-download-status").textContent="下载未完成，请稍后重试。";
  } finally {
    link?.remove();
    setTimeout(()=>{
      if (url) URL.revokeObjectURL(url);
      if (generation===replyDownloadGeneration&&authenticated&&csrf===token) {replyDownloadBusy=false;renderReplyDownload();}
    },1000);
  }
}
$("download-reply-statistics").addEventListener("click",downloadReplyStatistics);
