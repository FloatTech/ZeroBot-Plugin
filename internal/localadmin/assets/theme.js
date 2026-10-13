"use strict";
// Only the appearance preference is stored here; credentials stay in cookies.
(() => {
  const key="zerobot.localadmin.theme.v1", choices=new Set(["system","light","dark"]);
  let media=null;
  try {media=window.matchMedia("(prefers-color-scheme: dark)");} catch {}
  function readChoice() {
    try {const value=window.localStorage.getItem(key); return choices.has(value) ? value : "system";} catch {return "system";}
  }
  let choice=readChoice();
  function apply() {
    document.documentElement.dataset.theme=choice==="system" ? (media?.matches ? "dark" : "light") : choice;
    for (const select of document.querySelectorAll("[data-theme-select]")) select.value=choice;
  }
  apply();
  function connectControls() {
    for (const select of document.querySelectorAll("[data-theme-select]")) {
      select.addEventListener("change",()=>{
        if (!choices.has(select.value)) {apply(); return;}
        choice=select.value;
        try {window.localStorage.setItem(key,choice);} catch {}
        apply();
      });
    }
    apply();
  }
  if (document.readyState==="loading") document.addEventListener("DOMContentLoaded",connectControls,{once:true});
  else connectControls();
  if (media?.addEventListener) media.addEventListener("change",()=>{if (choice==="system") apply();});
  window.addEventListener("storage",event=>{if (event.key===key || event.key===null) {choice=readChoice(); apply();}});
})();
