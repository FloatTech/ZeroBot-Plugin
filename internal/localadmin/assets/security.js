"use strict";
let securityActive=false, securityBusy=false, securityGeneration=0, securityRequest=null;
const securityInputIDs=["current-admin-password","new-admin-password","confirm-admin-password"];
function clearSecurityInputs() {for (const id of securityInputIDs) $(id).value="";}
function securityControls() {
  $("security-fields").disabled=!authenticated || securityBusy;
  for (const button of document.querySelectorAll(".nav-button")) button.disabled=securityBusy;
  $("logout").disabled=securityBusy;
}
function resetSecurity() {
  securityGeneration++;
  if (securityRequest) securityRequest.abort();
  securityRequest=null;securityActive=false;securityBusy=false;
  clearSecurityInputs();$("security-error").hidden=true;$("security-status").textContent="";securityControls();
}
function showSecurity(active) {
  if (!active && securityActive) resetSecurity();
  securityActive=active;securityControls();
}
function securityCanLeave() {
  if (!securityBusy) return true;
  $("security-status").textContent="密码修改请求尚未结束，请等待结果后再切页或退出。";
  return false;
}
function otherSettingsBusy() {
  return connectionBusy || pluginSaving || pluginScopeBusy || (typeof aiBusy!=="undefined" && aiBusy) || (typeof backupBusy!=="undefined" && backupBusy);
}
async function changeAdminPassword(event) {
  event.preventDefault();
  if (!authenticated || !securityActive || securityBusy) return;
  $("security-error").hidden=true;
  const body={currentPassword:$(securityInputIDs[0]).value,newPassword:$(securityInputIDs[1]).value,confirmPassword:$(securityInputIDs[2]).value};
  if (!body.currentPassword || new TextEncoder().encode(body.currentPassword).length>512 || Array.from(body.newPassword).length<12 || new TextEncoder().encode(body.newPassword).length>512) {
    showError("security-error",new Error("请填写当前密码；新密码至少12个字符，且不超过512字节。"));return;
  }
  if (body.newPassword!==body.confirmPassword) {showError("security-error",new Error("两次输入的新密码不一致。"));return;}
  if (body.newPassword===body.currentPassword) {showError("security-error",new Error("新密码需与当前密码不同。"));return;}
  if (otherSettingsBusy()) {showError("security-error",new Error("其他设置请求尚未完成，请等待结果后再修改密码。"));return;}
  const generation=securityGeneration,requestToken=csrf;
  if (!window.confirm("修改成功后所有后台会话都将退出，未保存编辑会清空。是否修改后台密码？")) return;
  if (!authenticated || !securityActive || securityBusy || generation!==securityGeneration || csrf!==requestToken || otherSettingsBusy() || securityInputIDs.some((id,i)=>$(id).value!==[body.currentPassword,body.newPassword,body.confirmPassword][i])) return;
  const controller=new AbortController();securityRequest=controller;securityBusy=true;securityControls();
  $("security-status").textContent="正在保存新密码…";
  try {
    const value=await api("/api/password",body,controller.signal);
    if (generation!==securityGeneration || csrf!==requestToken || !authenticated) return;
    if (value.passwordChanged!==true || value.authenticated!==false) throw new Error("提交结果无法确认，请重新登录核对新密码。");
    authenticated=false;csrf="";showAuth();
    $("auth-notice").textContent="后台密码已修改，所有后台会话已退出。请使用新密码登录。";
  } catch(error) {
    if (generation===securityGeneration && csrf===requestToken && authenticated && error.name!=="AbortError") {
      showError("security-error",error);$("security-status").textContent="";
    }
  } finally {
    if (generation===securityGeneration && securityRequest===controller) {
      securityRequest=null;securityBusy=false;clearSecurityInputs();securityControls();
    }
  }
}
$("password-change-form").addEventListener("submit",changeAdminPassword);
$("clear-admin-passwords").addEventListener("click",()=>{if (!authenticated || !securityActive || securityBusy) return;clearSecurityInputs();$("security-error").hidden=true;$("security-status").textContent="输入已清空。";});
window.addEventListener("beforeunload",event=>{if (authenticated && securityBusy) {event.preventDefault();event.returnValue="";}});
window.addEventListener("pagehide",resetSecurity);
