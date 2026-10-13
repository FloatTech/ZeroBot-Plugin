"use strict";
let backupPreview = null, backupCleanupPreview = null, backupBusy = false, backupGeneration = 0, backupActive = false, backupRequest = null;
const backupScopeNames = {"startup-config":"启动配置", "control-db":"控制数据库", "admin-options":"后台启用与地址"};
let backupCatalog=null, backupSearch="", backupScopeFilter="all", backupStateFilter="all", backupReadFailed=false, backupReadAt=null;
let backupPage=1, backupPageSize=20, backupPageCount=1, backupMatchCount=0;
let backupSort="default";
let backupFromDate="", backupUntilDate="", backupDateStart=null, backupDateEnd=null, backupDateTimeZone="";
let backupExpiryTimer=null, backupExpiryGeneration=0;
const expiredBackupPreviews=new WeakSet();

function stopBackupExpiryTimer() {
  backupExpiryGeneration++;clearTimeout(backupExpiryTimer);backupExpiryTimer=null;
}
function validBackupExpiry(preview) {
  return Number.isSafeInteger(preview?.expiresAt) && preview.expiresAt>0 && Number.isFinite(new Date(preview.expiresAt*1000).getTime());
}
function backupPreviewExpired(preview) {
  return !validBackupExpiry(preview) || expiredBackupPreviews.has(preview) || preview.expiresAt*1000<=Date.now();
}
function renderBackupExpiry() {
  stopBackupExpiryTimer();let ticking=false;
  for (const [kind,preview,actionable] of [["restore",backupPreview,!!backupPreview?.changed],["cleanup",backupCleanupPreview,!!backupCleanupPreview?.candidates?.length]]) {
    const line=$(kind+"-expiry"),state=$(kind+"-expiry-state"),confirm=$(kind+"-confirm"),button=$(kind==="restore" ? "restore-backup" : "cleanup-backups");
    let message="",expired=false;
    if (!preview) {
      line.textContent="";confirm.checked=false;confirm.disabled=true;button.disabled=true;
    } else if (!authenticated || !backupActive) {
      confirm.disabled=true;button.disabled=true;
    } else if (backupBusy) {
      confirm.disabled=true;button.disabled=true;message="请求处理中，请等待结果；计时不会中断已提交的操作。";
    } else if (backupPreviewExpired(preview)) {
      expired=true;
      if (typeof preview==="object") expiredBackupPreviews.add(preview);confirm.checked=false;confirm.disabled=true;button.disabled=true;
      line.textContent=validBackupExpiry(preview) ? "预览有效至 "+new Date(preview.expiresAt*1000).toLocaleTimeString("zh-CN",{hour12:false})+"；剩余 0分00秒。" : "预览有效期无法确认。";
      message=validBackupExpiry(preview) ? "预览已过期，请重新预览并核对后再次确认。" : "无法确认有效期，请重新预览后再确认。";
    } else {
      const seconds=Math.ceil((preview.expiresAt*1000-Date.now())/1000);
      line.textContent="预览有效至 "+new Date(preview.expiresAt*1000).toLocaleTimeString("zh-CN",{hour12:false})+"；剩余 "+Math.floor(seconds/60)+"分"+String(seconds%60).padStart(2,"0")+"秒（本机估算）。";
      confirm.disabled=false;button.disabled=!actionable || !confirm.checked;
      message="到期需重新预览；提交时仍会校验当前状态，不会自动"+(kind==="restore" ? "恢复。" : "删除。");ticking=true;
    }
    state.dataset.expired=String(expired);
    if (state.textContent!==message) state.textContent=message;
  }
  if (ticking && authenticated && backupActive && !backupBusy && !document.hidden) {
    const generation=backupExpiryGeneration;
    backupExpiryTimer=setTimeout(()=>{if (generation===backupExpiryGeneration && authenticated && backupActive && !backupBusy && !document.hidden) renderBackupExpiry();},1000);
  }
}

function resetBackups() {
  backupGeneration++; backupRequest?.abort(); backupRequest = null;
  backupPreview = null; backupBusy = false; backupActive = false;
  backupCatalog=null;backupSearch="";backupScopeFilter="all";backupStateFilter="all";backupReadFailed=false;
  backupReadAt=null;$("backup-export-status").textContent="";
  backupFromDate="";backupUntilDate="";backupDateStart=null;backupDateEnd=null;backupDateTimeZone="";
  $("backup-from-date").value="";$("backup-until-date").value="";$("backup-date-error").hidden=true;
  renderBackupSizeSummary([],[]);
  backupPage=1;backupPageSize=20;backupPageCount=1;backupMatchCount=0;backupSort="default";$("backup-page-size").value="20";
  renderBackupPagination();
  $("backup-search").value="";$("backup-scope-filter").value="all";$("backup-state-filter").value="all";
  $("backup-filter-count").textContent="尚未读取备份列表。";$("restore-backup-id").textContent="";
  clearCleanupPreview();
  $("backup-list").replaceChildren(); $("restore-preview").hidden = true;
  $("restore-confirm").checked = false; $("restore-backup").disabled = true;
  $("restore-impact-list").replaceChildren();$("restore-impact-status").textContent="";
  $("backup-error").hidden = true; $("backup-status").textContent = "进入页面后读取备份。";
  $("reload-backups").disabled = false; $("create-backup").disabled = true;
  for (const id of ["backup-search","backup-scope-filter","backup-state-filter","clear-backup-filter"]) $(id).disabled=false;
  renderBackupDateFilter();
}
function showBackups(active) {
  backupActive = active;
  if (!active) {
    backupGeneration++; backupRequest?.abort(); backupRequest = null;
    backupBusy = false; clearRestorePreview();
    clearCleanupPreview();
    $("restore-confirm").checked = false; $("restore-backup").disabled = true;
    renderBackupExportControls();
    renderBackupDateFilter();
  } else loadBackups();
}
function backupReady(generation) { return generation === backupGeneration && authenticated && backupActive; }
function clearRestorePreview() {
  $("restore-impact-list").replaceChildren();$("restore-impact-status").textContent="";
  backupPreview = null; $("restore-preview").hidden = true;
  $("restore-backup-id").textContent="";
  $("restore-confirm").checked = false; $("restore-backup").disabled = true;
  renderBackupExpiry();
}
function renderRestoreImpacts(impacts) {
  const list=$("restore-impact-list");list.replaceChildren();
  const items=Array.isArray(impacts) ? impacts.filter(item=>item && typeof item.label==="string" && typeof item.changed==="boolean").slice(0,12) : [];
  $("restore-impact-status").textContent=items.length ? "共 "+items.length+" 个类别，"+items.filter(item=>item.changed).length+" 个有变化；按预览时当前保存状态比较。" : "当前预览未提供类别摘要，以上方覆盖说明为准。";
  for (const item of items) {
    const row=document.createElement("article"),label=document.createElement("strong"),state=document.createElement("span");
    row.className="restore-impact-row";label.textContent=item.label;
    state.className="restore-impact-state"+(item.changed ? " impact-changed" : "");state.textContent=item.changed ? "有变化" : "一致";
    row.append(label,state);
    if (Number.isSafeInteger(item.currentCount) && item.currentCount>=0 && Number.isSafeInteger(item.backupCount) && item.backupCount>=0) {
      const counts=document.createElement("p");counts.className="field-help";
      const unit=typeof item.unit==="string" ? item.unit : "";
      counts.textContent="当前 "+item.currentCount+unit+" → 备份 "+item.backupCount+unit+(item.changed && item.currentCount===item.backupCount ? "（数量相同，内容不同）" : "");row.append(counts);
    }
    list.append(row);
  }
}
function backupControls(busy) {
  if (busy) $("backup-export-status").textContent="";
  backupBusy = busy; $("reload-backups").disabled = busy;
  $("create-backup").disabled = busy || !$("backup-scope").value;
  $("backup-scope").disabled = busy;
  $("preview-cleanup").disabled = busy;
  $("cleanup-keep").disabled = busy; $("cleanup-days").disabled = busy;
  $("cleanup-backups").disabled = busy || !backupCleanupPreview?.candidates?.length || !$("cleanup-confirm").checked;
  $("restore-backup").disabled = busy || !backupPreview?.changed || !$("restore-confirm").checked;
  for (const id of ["backup-search","backup-scope-filter","backup-state-filter","clear-backup-filter"]) $(id).disabled=busy;
  for (const button of $("backup-list").querySelectorAll("button")) button.disabled = busy || button.dataset.invalid === "true";
  renderBackupPagination();
  renderBackupExpiry();
}
function renderBackupPagination() {
  renderBackupExportControls();
  renderBackupDateFilter();
  $("backup-page-status").textContent=backupMatchCount ? "第 "+backupPage+" / "+backupPageCount+" 页；显示 "+((backupPage-1)*backupPageSize+1)+"～"+Math.min(backupPage*backupPageSize,backupMatchCount)+" / "+backupMatchCount+" 份匹配记录。" : backupCatalog ? "没有匹配记录。" : "尚未读取备份。";
  $("backup-prev").disabled=backupBusy || !backupMatchCount || backupPage<=1;
  $("backup-next").disabled=backupBusy || !backupMatchCount || backupPage>=backupPageCount;
  $("backup-page-size").disabled=backupBusy;
  $("backup-sort").value=backupSort;
  $("backup-sort").disabled=backupBusy || !authenticated || !backupActive || !backupCatalog;
}
function renderBackups(value) {
  backupCatalog=value;backupReadFailed=false;backupReadAt=new Date().toISOString(); $("backup-scope").replaceChildren();
  for (const [scope, label] of Object.entries(backupScopeNames)) {
    if (!value.scopes?.[scope]) continue;
    const option = document.createElement("option"); option.value = scope; option.textContent = label; $("backup-scope").append(option);
  }
  $("backup-scope").value = $("backup-scope").firstElementChild?.value || "";
  renderBackupList();
  $("backup-status").textContent = value.more ? "仅列出最近 200 个备份。更早记录保留在本机目录。" : "已读取 " + (value.backups || []).length + " 个备份。恢复前会再次校验完整性和兼容性。";
}
function backupEntryState(entry) { return !entry.valid ? "invalid" : entry.integrity ? "verified" : "legacy"; }
function backupEntryScope(entry) { return Object.hasOwn(backupScopeNames,entry.scope) ? entry.scope : "unknown"; }
function filteredBackupEntries() {
  const items=(backupCatalog?.backups || []).filter(entry=>{
    const state=backupEntryState(entry),scope=backupEntryScope(entry);
    const label=scope==="unknown" ? "异常/未知范围" : backupScopeNames[scope];
    const date=entry.valid ? new Date(entry.createdAt).toLocaleString("zh-CN",{hour12:false}) : "";
    const time=entry.valid===true && typeof entry.createdAt==="string" ? Date.parse(entry.createdAt) : NaN;
    const dateMatches=backupDateStart===null && backupDateEnd===null || Number.isFinite(time) && (backupDateStart===null || time>=backupDateStart) && (backupDateEnd===null || time<backupDateEnd);
    return dateMatches && (backupScopeFilter==="all" || scope===backupScopeFilter) && (backupStateFilter==="all" || state===backupStateFilter) && [entry.id,label,entry.createdAt || "",date].join(" ").toLocaleLowerCase().includes(backupSearch);
  });
  return sortBackupEntries(items);
}
function sortBackupEntries(items) {
  if (backupSort==="default") return items;
  const key=entry=>{
    if (entry.valid!==true) return null;
    if (backupSort==="oldest") {
      const value=typeof entry.createdAt==="string" ? Date.parse(entry.createdAt) : NaN;
      return Number.isFinite(value) ? value : null;
    }
    return Number.isSafeInteger(entry.size) && entry.size>=0 ? entry.size : null;
  };
  return items.map((entry,index)=>({entry,index,key:key(entry)})).sort((a,b)=>{
    if (a.key===null || b.key===null) return (a.key===null)-(b.key===null) || a.index-b.index;
    return (backupSort==="largest" ? b.key-a.key : a.key-b.key) || a.index-b.index;
  }).map(item=>item.entry);
}
function renderBackupList() {
  $("backup-export-status").textContent="";
  $("backup-list").replaceChildren();
  const entries=backupCatalog?.backups || [],counts={verified:0,legacy:0,invalid:0};
  for (const entry of entries) counts[backupEntryState(entry)]++;
  const items=filteredBackupEntries();
  renderBackupSizeSummary(entries,items);
  $("backup-filter-count").textContent=backupCatalog ? items.length+" / "+entries.length+" 份已加载备份；含摘要 "+counts.verified+"，旧版无摘要 "+counts.legacy+"，不可恢复 "+counts.invalid+"。"+(backupCatalog.more ? " 仅检索本次加载的最新 200 份，更早未加载记录不在结果中。" : "")+(backupReadFailed ? " 上次读取结果，刷新失败。" : "") : "尚未读取备份列表。";
  backupMatchCount=items.length;backupPageCount=Math.max(1,Math.ceil(items.length/backupPageSize));backupPage=Math.min(backupPage,backupPageCount);
  renderBackupPagination();
  for (const entry of items.slice((backupPage-1)*backupPageSize,backupPage*backupPageSize)) {
    const card = document.createElement("article"), info = document.createElement("div"), title = document.createElement("h2"), detail = document.createElement("p"), id = document.createElement("small"), button = document.createElement("button");
    card.className = "panel backup-card";
    title.textContent = backupScopeNames[entry.scope] || "无效备份";
    detail.className = "muted";
    detail.textContent = entry.valid ? new Date(entry.createdAt).toLocaleString("zh-CN", {hour12:false}) + " · " + Math.ceil(entry.size/1024) + " KiB · 格式 v" + entry.version + (entry.integrity ? " · 含校验摘要" : " · 旧版无摘要") : "损坏或格式不支持，不能恢复。";
    id.textContent = entry.id; button.className = "secondary"; button.textContent = "预览恢复";
    button.disabled = backupBusy || !entry.valid; button.dataset.invalid = String(!entry.valid);
    button.addEventListener("click", () => previewBackup(entry.id));
    info.append(title,detail,id); card.append(info,button); $("backup-list").append(card);
  }
  if (!items.length) {
    const empty = document.createElement("p"); empty.className = "empty";
    empty.textContent = !backupCatalog ? "尚未读取备份，请刷新列表。" : entries.length ? "没有符合筛选条件的备份，请调整条件或清除筛选。" : "暂无备份。可以手动创建，配置修改前也会自动备份。";
    $("backup-list").append(empty);
  }
}
function backupSizeStats(entries) {
  let bytes=0n,unknown=0;
  for (const entry of entries) {
    if (entry.valid===true && Number.isSafeInteger(entry.size) && entry.size>=0) bytes+=BigInt(entry.size);
    else unknown++;
  }
  return {count:entries.length,bytes,unknown};
}
function formatBackupBytes(bytes) {
  if (bytes<1024n) return bytes+" 字节";
  const units=["字节","KiB","MiB","GiB","TiB","PiB","EiB"];
  let unit=0,scale=1n;
  while (unit<units.length-1 && bytes>=scale*1024n) {unit++;scale*=1024n;}
  const rounded=(bytes*100n+scale/2n)/scale;
  return rounded/100n+"."+String(rounded%100n).padStart(2,"0")+" "+units[unit]+"（"+bytes+" 字节）";
}
function backupSizeDescription(entries) {
  const stats=backupSizeStats(entries);
  return stats.count+" 份；可核对内容大小："+formatBackupBytes(stats.bytes)+"；大小无法核对："+stats.unknown+" 份。";
}
function renderBackupSizeSummary(entries,items) {
  const panel=$("backup-size-summary"),list=$("backup-size-scopes");
  panel.hidden=!backupCatalog;list.replaceChildren();
  for (const id of ["backup-size-loaded","backup-size-matched","backup-size-status"]) $(id).textContent="";
  if (!backupCatalog) return;
  $("backup-size-loaded").textContent="已加载："+backupSizeDescription(entries);
  $("backup-size-matched").textContent="当前匹配（跨全部已加载页）："+backupSizeDescription(items);
  $("backup-size-status").textContent="最近成功读取："+new Date(backupReadAt).toLocaleString("zh-CN",{hour12:false})+"。"+(backupReadFailed ? "上次读取结果，刷新失败。" : "")+(backupCatalog.more ? "仅加载最新200份，更早未加载记录不在摘要中。" : "");
  for (const [scope,label] of [...Object.entries(backupScopeNames),["unknown","异常/未知范围"]]) {
    const row=document.createElement("article"),title=document.createElement("h3"),loaded=document.createElement("p"),matched=document.createElement("p");
    title.textContent=label;loaded.textContent="已加载："+backupSizeDescription(entries.filter(entry=>backupEntryScope(entry)===scope));
    matched.textContent="当前匹配："+backupSizeDescription(items.filter(entry=>backupEntryScope(entry)===scope));
    row.append(title,loaded,matched);list.append(row);
  }
}
function renderBackupExportControls() {
  $("download-backup-catalog").disabled=!authenticated || !backupActive || backupBusy || !backupCatalog;
}
function downloadBackupCatalog() {
  if (!authenticated || !backupActive || backupBusy || !backupCatalog) return;
  let url=null,link=null;
  try {
    const exportedAt=new Date().toISOString(),entries=filteredBackupEntries().map(entry=>({
      id:typeof entry.id==="string" && /^[0-9]{8}T[0-9]{6}\.[0-9]{9}Z-[0-9a-f]{16}(\.json|-control)$/.test(entry.id) ? entry.id : null,
      createdAt:entry.valid && typeof entry.createdAt==="string" && Number.isFinite(Date.parse(entry.createdAt)) ? new Date(entry.createdAt).toISOString() : null,
      scope:backupEntryScope(entry),
      size:Number.isSafeInteger(entry.size) && entry.size>=0 ? entry.size : null,
      version:Number.isSafeInteger(entry.version) && entry.version>0 ? entry.version : null,
      integrity:entry.integrity===true,valid:entry.valid===true
    }));
    // Explicit metadata fields only: never serialize the catalog or restore ticket.
    const value={format:"zerobot-backup-catalog",version:1,exportedAt,readAt:backupReadAt,
      source:{loaded:backupCatalog.backups.length,matched:entries.length,more:backupCatalog.more===true,lastReadFailed:backupReadFailed},
      filters:{scope:backupScopeFilter,state:backupStateFilter,keywordApplied:!!backupSearch,sort:backupSort,
        ...(backupFromDate || backupUntilDate ? {fromDate:backupFromDate,untilDate:backupUntilDate,dateTimeZone:backupDateTimeZone} : {})},backups:entries};
    url=URL.createObjectURL(new Blob([JSON.stringify(value,null,2)+"\n"],{type:"application/json;charset=utf-8"}));
    link=document.createElement("a");link.href=url;link.download="zerobot-backup-catalog-"+exportedAt.replace(/[-:.]/g,"")+".json";
    document.body.append(link);link.click();
    $("backup-export-status").textContent="已发起清单下载，共 "+entries.length+" 份匹配记录（跨全部已加载页）。"+(backupReadFailed ? " 使用上次读取结果，最近刷新失败。" : "")+(backupCatalog.more ? " 更早未加载备份未列入。" : "")+" 清单仅供核对，不能用于恢复。";
  } catch {
    $("backup-export-status").textContent="清单下载未完成，请检查浏览器下载设置后重试。";
  } finally {
    link?.remove();if (url) window.setTimeout(()=>URL.revokeObjectURL(url),1000);
  }
}
$("download-backup-catalog").addEventListener("click",downloadBackupCatalog);
$("backup-sort").addEventListener("change",()=>{
  const sort=$("backup-sort").value;
  if (!authenticated || !backupActive || backupBusy || !backupCatalog || !["default","oldest","largest","smallest"].includes(sort)) {$("backup-sort").value=backupSort;return;}
  backupSort=sort;backupPage=1;renderBackupList();
});
function backupDateBoundary(value, nextDay=false) {
  if (!value) return null;
  const match=/^(\d{4})-(\d{2})-(\d{2})$/.exec(value);
  if (!match || Number(match[1])<1) throw new Error("请填写有效的本地日期。");
  const [year,month,day]=match.slice(1).map(Number),date=new Date(0);
  date.setFullYear(year,month-1,day);date.setHours(0,0,0,0);
  if (!Number.isFinite(date.getTime()) || date.getFullYear()!==year || date.getMonth()!==month-1 || date.getDate()!==day) throw new Error("请填写有效的本地日期。");
  // Use the next calendar day, not a fixed 24 hours, across daylight-saving changes.
  if (nextDay) {date.setDate(date.getDate()+1);date.setHours(0,0,0,0);}
  return date.getTime();
}
function renderBackupDateFilter() {
  const draft=$("backup-from-date").value!==backupFromDate || $("backup-until-date").value!==backupUntilDate;
  $("backup-date-status").textContent=(backupFromDate || backupUntilDate ? "已应用日期（本地，包含结束当天）："+(backupFromDate || "不限")+" 至 "+(backupUntilDate || "不限")+"。有日期限制时排除无法核对创建时间的记录。" : "已应用：不限日期。")+(draft ? " 日期修改尚未应用，列表、摘要与清单仍使用原条件。" : "");
  const disabled=backupBusy || !authenticated || !backupActive || !backupCatalog;
  for (const id of ["backup-from-date","backup-until-date","apply-backup-date"]) $(id).disabled=disabled;
}
function applyBackupDateFilter() {
  if (!authenticated || !backupActive || backupBusy || !backupCatalog) return;
  const from=$("backup-from-date").value,until=$("backup-until-date").value;
  try {
    if (!$("backup-from-date").validity.valid || !$("backup-until-date").validity.valid) throw new Error("请填写有效的本地日期。");
    const start=backupDateBoundary(from),end=backupDateBoundary(until,true);
    if (from && until && from>until) throw new Error("开始日期不能晚于结束日期。");
    backupFromDate=from;backupUntilDate=until;backupDateStart=start;backupDateEnd=end;
    backupDateTimeZone=Intl.DateTimeFormat().resolvedOptions().timeZone;
    $("backup-date-error").hidden=true;backupPage=1;renderBackupList();
  } catch(error) {showError("backup-date-error",error);renderBackupDateFilter();}
}
$("backup-date-filter").addEventListener("submit",event=>{event.preventDefault();applyBackupDateFilter();});
for (const id of ["backup-from-date","backup-until-date"]) $(id).addEventListener("input",()=>{$("backup-date-error").hidden=true;renderBackupDateFilter();});
function applyBackupFilter() {
  if (!authenticated || !backupActive || backupBusy) return;
  const search=$("backup-search").value.toLocaleLowerCase().trim(),scope=$("backup-scope-filter").value || "all",state=$("backup-state-filter").value || "all";
  if (search!==backupSearch || scope!==backupScopeFilter || state!==backupStateFilter) backupPage=1;
  backupSearch=search;backupScopeFilter=scope;backupStateFilter=state;
  renderBackupList();
}
$("backup-filter").addEventListener("submit",event=>{event.preventDefault();applyBackupFilter();});
$("backup-search").addEventListener("input",applyBackupFilter);
for (const id of ["backup-scope-filter","backup-state-filter"]) $(id).addEventListener("change",applyBackupFilter);
$("clear-backup-filter").addEventListener("click",()=>{
  if (!authenticated || !backupActive || backupBusy) return;
  backupFromDate="";backupUntilDate="";backupDateStart=null;backupDateEnd=null;backupDateTimeZone="";
  $("backup-from-date").value="";$("backup-until-date").value="";$("backup-date-error").hidden=true;
  backupPage=1;$("backup-search").value="";$("backup-scope-filter").value="all";$("backup-state-filter").value="all";applyBackupFilter();
});
function changeBackupPage(step) {
  if (!authenticated || !backupActive || backupBusy) return;
  const next=backupPage+step;if (next<1 || next>backupPageCount) return;
  backupPage=next;renderBackupList();
}
$("backup-prev").addEventListener("click",()=>changeBackupPage(-1));
$("backup-next").addEventListener("click",()=>changeBackupPage(1));
$("backup-page-size").addEventListener("change",()=>{
  if (!authenticated || !backupActive || backupBusy) return;
  const size=Number($("backup-page-size").value);
  if (![10,20,50].includes(size)) {$("backup-page-size").value=String(backupPageSize);return;}
  backupPageSize=size;backupPage=1;renderBackupList();
});
async function loadBackups() {
  if (!authenticated || !backupActive || backupBusy) return;
  const generation = ++backupGeneration; clearRestorePreview(); clearCleanupPreview(); $("backup-error").hidden = true; backupControls(true);
  const controller = new AbortController(); backupRequest = controller;
  try { const value = await api("/api/backups", undefined, controller.signal); if (backupReady(generation)) renderBackups(value); }
  catch (error) { if (backupReady(generation) && error.name !== "AbortError") {backupReadFailed=true;renderBackupList();$("backup-status").textContent=backupCatalog ? "备份列表读取失败；当前显示上次读取结果。" : "备份列表读取失败；尚未取得记录。";showError("backup-error",error);} }
  finally { if (backupReady(generation)) {backupRequest=null; backupControls(false);} }
}
async function previewBackup(id) {
  if (!authenticated || !backupActive || backupBusy) return;
  const generation = ++backupGeneration; clearRestorePreview(); clearCleanupPreview(); $("backup-error").hidden = true; backupControls(true);
  const controller = new AbortController(); backupRequest = controller;
  try {
    const value = await api("/api/backups/preview", {id}, controller.signal);
    if (!backupReady(generation)) return;
    backupPreview = value; $("restore-preview").hidden = false;
    $("restore-backup-id").textContent="备份编号："+value.backup.id;
    $("restore-title").textContent = "恢复预览：" + backupScopeNames[value.backup.scope];
    $("restore-detail").textContent = value.changed ? "此操作将覆盖该范围的当前状态。" + (value.backup.scope === "control-db" ? " 包含 " + value.tables + " 个表、" + value.rows + " 条记录，" + value.changedTables + " 个表有变化。" : "") : "备份与当前配置一致，无需恢复。";
    renderRestoreImpacts(value.impacts);
    $("restore-warnings").replaceChildren();
    for (const warning of value.warnings || []) {const item=document.createElement("li"); item.textContent=warning; $("restore-warnings").append(item);}
    $("restore-preview").scrollIntoView?.({block:"start",behavior:window.matchMedia?.("(prefers-reduced-motion: reduce)").matches ? "instant" : "smooth"});
  } catch (error) { if (backupReady(generation) && error.name !== "AbortError") showError("backup-error",error); }
  finally { if (backupReady(generation)) {backupRequest=null; backupControls(false);} }
}
$("reload-backups").addEventListener("click",loadBackups);
$("backup-scope").addEventListener("change",()=>backupControls(backupBusy));
$("restore-confirm").addEventListener("change",()=>backupControls(backupBusy));
$("cancel-restore").addEventListener("click",clearRestorePreview);
$("backup-create-form").addEventListener("submit",async event=>{
  event.preventDefault(); if (!authenticated || !backupActive || backupBusy || !$("backup-scope").value) return;
  const generation=++backupGeneration, scope=$("backup-scope").value;
  clearRestorePreview(); clearCleanupPreview(); $("backup-error").hidden=true; backupControls(true);
  try {
    await api("/api/backups",{scope});
    if (!backupReady(generation)) return;
    backupControls(false); await loadBackups();
    if (authenticated && backupActive) $("backup-status").textContent += " 手动备份已创建。";
  } catch(error) { if (backupReady(generation)) showError("backup-error",new Error(error.message + " 请求中断时，请刷新列表确认是否已创建；不要连续重复提交。")); }
  finally { if (backupReady(generation)) backupControls(false); }
});
$("restore-backup").addEventListener("click",async()=>{
  if (!authenticated || !backupActive || backupBusy || !backupPreview?.changed || !$("restore-confirm").checked) return;
  if (backupPreviewExpired(backupPreview)) {clearRestorePreview(); showError("backup-error",new Error("预览已过期或有效期无法确认，请重新预览。")); return;}
  const generation=++backupGeneration, preview=backupPreview;
  $("backup-error").hidden=true; backupControls(true);
  try {
    const result=await api("/api/backups/restore",{id:preview.backup.id,revision:preview.revision,expiresAt:preview.expiresAt,token:preview.token,confirm:true});
    if (!backupReady(generation)) return;
    // Restored values must be reloaded in their own pages; no stale drafts survive.
    if (typeof resetConnections === "function") resetConnections(); else connectionSettings=null;
    if (typeof resetAI === "function") resetAI();
    plugins=[]; pluginDrafts.clear(); pluginListRequest++; pluginGeneration++;
    pluginSaving=false; pluginScopeBusy=false; pluginBatch=null; pluginControlAvailable=false;
    pluginResponse=null; pluginResponseLoaded=false;
    pluginReloadAfterRestore=true;
    $("plugin-save-state").textContent=""; $("plugin-batch-status").textContent="";
    resetPluginCatalog(); renderPluginBusy(); renderPlugins();
    clearRestorePreview(); backupControls(false); await loadBackups();
    if (authenticated && backupActive) $("backup-status").textContent = result.changed ? "恢复完成，恢复前状态已备份。" + (result.restartRequired ? " 请重启机器人完成全部生效。" : "") : "配置一致，无需恢复。";
  } catch(error) { if (backupReady(generation)) {clearRestorePreview(); showError("backup-error",new Error(error.message + " 如果请求中断，请检查当前配置后再重新预览。"));} }
  finally { if (backupReady(generation)) backupControls(false); }
});

function clearCleanupPreview() {
  backupCleanupPreview=null; $("cleanup-preview").hidden=true;
  $("cleanup-confirm").checked=false; $("cleanup-backups").disabled=true;
  $("cleanup-list").replaceChildren();
  renderBackupExpiry();
}
$("cleanup-confirm").addEventListener("change",()=>backupControls(backupBusy));
$("cancel-cleanup").addEventListener("click",clearCleanupPreview);
for (const id of ["cleanup-keep","cleanup-days"]) $(id).addEventListener("input",()=>{if (!backupBusy) {clearCleanupPreview();backupControls(false);}});
$("cleanup-form").addEventListener("submit",async event=>{
  event.preventDefault(); if (!authenticated || !backupActive || backupBusy) return;
  const policy={keepPerScope:Number($("cleanup-keep").value),keepDays:Number($("cleanup-days").value)};
  if (!Number.isInteger(policy.keepPerScope) || policy.keepPerScope<3 || policy.keepPerScope>200 || !Number.isInteger(policy.keepDays) || policy.keepDays<1 || policy.keepDays>3650) {
    clearCleanupPreview(); showError("backup-error",new Error("保留份数须为 3～200，保留天数须为 1～3650。")); return;
  }
  const generation=++backupGeneration;clearRestorePreview();clearCleanupPreview();$("backup-error").hidden=true;backupControls(true);
  const controller=new AbortController();backupRequest=controller;
  try {
    const value=await api("/api/backups/cleanup-preview",policy,controller.signal);
    if (!backupReady(generation)) return;
    backupCleanupPreview=value;$("cleanup-preview").hidden=false;
    $("cleanup-detail").textContent="本次将删除 "+value.candidates.length+" 份旧备份，约 "+Math.ceil(value.bytes/1024)+" KiB；跳过 "+value.skipped+" 份无法安全确认的备份。"+(value.more ? " 每次最多 100 份，剩余部分需重新预览。" : "")+" 永久删除不能撤销，当前配置与对话记录不受影响。";
    for (const entry of value.candidates) {const item=document.createElement("li");item.textContent=(backupScopeNames[entry.scope] || entry.scope)+" · "+new Date(entry.createdAt).toLocaleString("zh-CN",{hour12:false})+" · "+entry.id;$("cleanup-list").append(item);}
  } catch(error) {if (backupReady(generation) && error.name!=="AbortError") showError("backup-error",error);}
  finally {if (backupReady(generation)) {backupRequest=null;backupControls(false);}}
});
$("cleanup-backups").addEventListener("click",async()=>{
  if (!authenticated || !backupActive || backupBusy || !backupCleanupPreview?.candidates?.length || !$("cleanup-confirm").checked) return;
  const preview=backupCleanupPreview;
  if (backupPreviewExpired(preview)) {clearCleanupPreview();showError("backup-error",new Error("清理预览已过期或有效期无法确认，请重新预览。"));return;}
  const generation=++backupGeneration;$("backup-error").hidden=true;backupControls(true);
  try {
    const result=await api("/api/backups/cleanup",{policy:preview.policy,revision:preview.revision,asOf:preview.asOf,expiresAt:preview.expiresAt,token:preview.token,confirm:true});
    if (!backupReady(generation)) return;
    clearCleanupPreview();backupControls(false);const refreshGeneration=backupGeneration+1;await loadBackups();
    if (backupReady(refreshGeneration)) {$("backup-status").textContent=result.message+" 已完整删除 "+result.deleted.length+" 份。";if(result.partial) showError("backup-error",new Error(result.message));}
  } catch(error) {if (backupReady(generation)) {clearCleanupPreview();showError("backup-error",new Error(error.message+" 请求中断时请刷新列表核对实际结果，已删除内容不会自动恢复。"));}}
  finally {if (backupReady(generation)) backupControls(false);}
});
document.addEventListener("visibilitychange",()=>{
  stopBackupExpiryTimer();
  if (!document.hidden && authenticated && backupActive) renderBackupExpiry();
});
