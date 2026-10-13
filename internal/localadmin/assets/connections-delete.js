"use strict";
let connectionDeletion = null, connectionDeleting = false;
function clearConnectionDeletion() {
  connectionDeletion=null;
  $("connection-delete-preview").hidden=true;
  $("connection-delete-confirm").checked=false;
  $("connection-delete-save").disabled=true;
  for(const id of ["connection-delete-target","connection-delete-address","connection-delete-impact"]) $(id).textContent="";
}
function resetConnectionDelete() {
  connectionDeleting=false;clearConnectionDeletion();
  $("connection-delete-status").textContent="";$("connection-delete-error").hidden=true;
}
function deletionCurrent(preview) {
  return preview && authenticated && preview.token===csrf && preview.generation===connectionGeneration && preview.settings===connectionSettings && preview.selected===connectionIndex && !connectionDirty();
}
function renderConnectionDeleteControls() {
  if(connectionDeletion && !deletionCurrent(connectionDeletion)) clearConnectionDeletion();
  const total=(connectionSettings?.connections || []).length;
  const blocked=!authenticated || connectionBusy || !connectionSettings?.available;
  $("preview-connection-delete").disabled=blocked || !selectedConnection() || total<=1 || connectionDirty();
  $("connection-delete-hint").textContent=!connectionSettings?.available ? "读取配置后可管理删除。" : total<=1 ? "至少保留一条连接，不能删除最后一条。" : connectionDirty() ? "请先保存或放弃现有及新增草稿，再预览删除。" : "预览只核对目标；确认删除后需重启机器人生效。";
  $("connection-delete-confirm").disabled=blocked || !connectionDeletion;
  $("connection-delete-save").disabled=blocked || !connectionDeletion || !$("connection-delete-confirm").checked;
  $("connection-delete-cancel").disabled=connectionBusy;
}
function showConnectionDelete(active) {
  if(!active) {clearConnectionDeletion();renderConnectionDeleteControls();}
}
function connectionDeleteCanLeave() {
  if(!connectionDeleting) return true;
  $("connection-delete-status").textContent="删除请求尚未完成，请稍候再切换页面。";
  return false;
}
$("preview-connection-delete").addEventListener("click",()=>{
  if(!authenticated || connectionBusy || !connectionSettings?.available || connectionDirty() || !selectedConnection() || connectionSettings.connections.length<=1) return;
  clearConnectionDeletion();$("connection-delete-error").hidden=true;
  const connection=selectedConnection();
  connectionDeletion={settings:connectionSettings,generation:connectionGeneration,token:csrf,selected:connectionIndex,kind:connection.kind,index:connection.index};
  $("connection-delete-target").textContent="删除目标："+(connection.kind==="ws" ? "正向连接 " : "反向监听 ")+(connection.index+1);
  $("connection-delete-address").textContent="已保存地址："+connection.url;
  $("connection-delete-impact").textContent="已保存连接共 "+connectionSettings.connections.length+" 条，删除后保留 "+(connectionSettings.connections.length-1)+" 条。同类型序号会调整，本次启动信息仍按原序号展示。";
  $("connection-delete-preview").hidden=false;$("connection-delete-status").textContent="请核对固定目标，预览尚未修改配置。";
  renderConnectionDeleteControls();
});
$("connection-delete-confirm").addEventListener("change",renderConnectionDeleteControls);
$("connection-delete-cancel").addEventListener("click",()=>{
  if(connectionBusy) return;
  clearConnectionDeletion();renderConnectionDeleteControls();$("connection-delete-status").textContent="已取消删除，配置保持不变。";
});
$("connection-delete-save").addEventListener("click",async()=>{
  const preview=connectionDeletion;
  if(connectionBusy || !deletionCurrent(preview) || !$("connection-delete-confirm").checked) {renderConnectionDeleteControls();return;}
  connectionDeleting=true;connectionControls(true);$("connection-delete-error").hidden=true;$("connection-delete-status").textContent="正在备份并删除所选连接…";
  try {
    const value=await api("/api/connections",{revision:preview.settings.revision,removeConnection:{kind:preview.kind,index:preview.index,confirm:true}});
    if(!deletionCurrent(preview) || preview!==connectionDeletion) return;
    if(!value.available || !value.backupId || !value.revision || !Array.isArray(value.connections) || value.connections.length!==preview.settings.connections.length-1) throw new Error("删除结果无法确认，请先重新读取配置核对。");
    connectionIndex=Math.min(preview.selected,value.connections.length-1);clearConnectionDeletion();renderConnectionSettings(value);
    $("connection-save-state").textContent+=" 删除前配置备份已创建。";$("connection-delete-status").textContent=value.restartRequired ? "所选连接已从保存配置删除，重启后生效；列表序号已调整，当前运行连接保持不变。" : "所选连接已从保存配置删除；已保存配置与本次启动一致，无需为本次删除重启。列表序号已调整。";
  } catch(error) {
    if(authenticated && preview.token===csrf && preview.generation===connectionGeneration) {
      clearConnectionDeletion();showError("connection-delete-error",error);
      $("connection-delete-status").textContent="删除未确认成功，确认已清空；若网络中断，请先重新读取配置核对。";
    }
  } finally {
    if(authenticated && preview.token===csrf && preview.generation===connectionGeneration) {connectionDeleting=false;connectionControls(false);renderConnectionDrafts();}
  }
});
