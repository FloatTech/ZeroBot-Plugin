"use strict";
const $ = (id) => document.getElementById(id);
let csrf = "", setupRequired = false, plugins = [], authenticated = false;
let connectionSettings = null, connectionBusy = false, connectionGeneration = 0, connectionIndex = 0, connectionBaseline = null;
let connectionDraftConfirmation = null;
let pluginGroup = "0", pluginControlAvailable = false, pluginSaving = false, pluginScopeBusy = false, pluginListRequest = 0, pluginGeneration = 0;
const pluginDrafts = new Map();
let pluginPage = 1, pluginPageSize = 20, pluginPageCount = 1, pluginMatchCount = 0, pluginSearchQuery = "", pluginStateFilter = "all";
let pluginOnlyDrafts = false;
let pluginSort = "default";
let pluginBatch = null;
let pluginResponse = null, pluginResponseLoaded = false;
let pluginUsage = null;
let pluginCatalogReadAt=null, pluginCatalogStateAt=null, pluginCatalogReadFailed=false, pluginCatalogLoading=null;
let pluginReloadAfterRestore=false;

async function api(path, body, signal) {
  const requestToken = csrf;
  const options = { credentials: "same-origin", cache: "no-store", signal };
  if (body !== undefined) {
    options.method = "POST";
    options.headers = { "Content-Type": "application/json", "X-Local-Request": "1", "X-CSRF-Token": csrf };
    options.body = JSON.stringify(body);
  }
  const response = await fetch(path, options);
  const value = await response.json();
  if (signal?.aborted) throw Object.assign(new Error("请求已取消"), {name:"AbortError"});
  if (!response.ok) {
    if (response.status === 401 && authenticated && csrf === requestToken) { authenticated = false; csrf = ""; showAuth(); }
    throw new Error(value.error || "请求未完成，请重试。");
  }
  return value;
}
function showError(id, error) { $(id).textContent = error.message; $(id).hidden = false; }
function showAuth() {
  closePluginUsage();
  if (typeof resetSecurity === "function") resetSecurity();
	if (typeof resetOverview === "function") resetOverview();
	if (typeof resetAI === "function") resetAI();
	if (typeof resetLogs === "function") resetLogs();
	if (typeof resetBackups === "function") resetBackups();
  resetConnections();
  pluginDrafts.clear(); pluginListRequest++; pluginGeneration++; plugins = [];
  pluginBatch = null; $("plugin-batch-status").textContent = "";
  pluginResponse = null; pluginResponseLoaded = false;
  pluginReloadAfterRestore = false;
  pluginGroup = "0"; pluginControlAvailable = false; pluginSaving = false; pluginScopeBusy = false;
  resetPluginCatalog();
  $("plugin-scope").value = "global"; $("plugin-group").value = "";
  $("plugin-error").hidden = true; $("plugin-save-state").textContent = "";
  $("plugin-search").value = ""; $("plugin-state-filter").value = "all";
  pluginPage = 1; pluginPageSize = 20; pluginSearchQuery = ""; pluginStateFilter = "all"; $("plugin-page-size").value = "20";
  pluginOnlyDrafts = false; $("plugin-only-drafts").checked = false;
  pluginSort = "default";
  renderPluginScope(); renderPluginBusy(); renderPlugins();
  $("auth").hidden = false; $("workspace").hidden = true;
  $("setup-fields").hidden = !setupRequired; $("confirm-fields").hidden = !setupRequired;
  $("setup-code").required = setupRequired; $("confirm-password").required = setupRequired;
  $("auth-title").textContent = setupRequired ? "建立后台密码" : "登录管理后台";
  $("auth-description").textContent = setupRequired ? "只需设置一次，用于管理这台电脑上的机器人。" : "使用本机后台密码继续。";
  $("password").autocomplete = setupRequired ? "new-password" : "current-password";
  $("password").minLength = setupRequired ? 12 : 1;
  $("login-button").textContent = setupRequired ? "设置密码并进入 →" : "登录后台 →";
}
async function showWorkspace() {
  authenticated = true; setupRequired = false;
  $("auth-notice").textContent = "";
  const pluginSession = pluginGeneration;
  $("auth").hidden = true; $("workspace").hidden = false; $("auth-error").hidden = true;
  $("password").value = ""; $("confirm-password").value = ""; $("setup-code").value = ""; $("setup-file").value = "";
  const view = document.querySelector(".nav-button.active")?.dataset.view;
  showOverview(view === "overview");
  try { await loadPluginList(); }
  catch (error) { if (authenticated && pluginSession === pluginGeneration) showError("plugin-error",error); }
  if (!authenticated || pluginSession !== pluginGeneration) return;
  const currentView = document.querySelector(".nav-button.active")?.dataset.view;
  if (typeof showLogs === "function") showLogs(currentView === "logs");
  if (typeof showBackups === "function") showBackups(currentView === "backups");
  if (typeof showSecurity === "function") showSecurity(currentView === "security");
  if (currentView === "connections") loadConnections();
  if (currentView === "ai") loadAI();
}
function duration(seconds) {
  if (seconds >= 86400) return Math.floor(seconds / 86400) + " 天";
  if (seconds >= 3600) return Math.floor(seconds / 3600) + " 小时 " + Math.floor(seconds % 3600 / 60) + " 分";
  return Math.floor(seconds / 60) + " 分钟";
}
function renderOverview(value) {
  renderOverviewAlerts(value.alerts);
  renderModelRequests(value.modelRequests);
  if (typeof renderReplyTasks==="function") renderReplyTasks(value.replyTasks);
  const accounts = value.botAccounts || [];
  $("account-count").textContent = accounts.length;
  $("plugin-count").textContent = value.pluginCount;
  $("nav-count").textContent = value.pluginCount;
  $("uptime").textContent = duration(value.uptimeSeconds);
  $("version").textContent = value.version;
  $("go-version").textContent = value.goVersion;
  $("configured-count").textContent = value.configuredConnections;
  $("checked-at").textContent = new Date(value.checkedAt).toLocaleTimeString("zh-CN", { hour12: false });
  const connected = value.connectionState === "connected";
  $("connection-badge").textContent = connected ? "已连接" : value.connectionState === "unconfigured" ? "未配置连接" : "等待连接";
  $("connection-badge").classList.toggle("waiting", !connected);
  $("connection-detail").textContent = connected ? "后台运行正常，已接入 OneBot。" : "后台运行正常，目前没有已连接的机器人账号。";
  $("accounts").replaceChildren();
  if (!accounts.length) {
    const empty = document.createElement("p"); empty.className = "empty"; empty.textContent = "暂时没有已连接的账号。"; $("accounts").append(empty);
  }
  for (const account of accounts) {
    const row = document.createElement("div"); row.className = "account-row";
    const avatar = document.createElement("span"); avatar.className = "account-avatar"; avatar.textContent = "QQ";
    const info = document.createElement("div"), name = document.createElement("strong"), note = document.createElement("small"), dot = document.createElement("span");
    name.textContent = account; note.textContent = "OneBot 已连接"; dot.className = "dot";
    info.append(name, note); row.append(avatar, info, dot); $("accounts").append(row);
  }
}
function sortPluginItems(items) {
  if (pluginSort === "default") return items;
  const rank = plugin => {
    if (pluginSort === "drafts") return pluginDrafts.has(plugin.name) ? 0 : 1;
    const state = pluginEffectiveState(plugin);
    return state === "unknown" ? 2 : state === pluginSort ? 0 : 1;
  };
  // Sort a view, retaining the loaded order for ties and sequential saves.
  return items.map((plugin,index) => ({plugin,index})).sort((a,b) => {
    const order = pluginSort === "name" ? a.plugin.name.localeCompare(b.plugin.name,"zh-CN",{numeric:true,sensitivity:"base"}) : rank(a.plugin)-rank(b.plugin);
    return order || a.index-b.index;
  }).map(item => item.plugin);
}
function filteredPluginItems() {
  const query=$("plugin-search").value.toLocaleLowerCase().trim(),filter=$("plugin-state-filter").value;
  return sortPluginItems(plugins.filter(p=>(!pluginOnlyDrafts || pluginDrafts.has(p.name)) && (filter==="all" || filter===pluginEffectiveState(p)) && [p.name, p.brief, p.usage].join(" ").toLocaleLowerCase().includes(query)));
}
function resetPluginCatalog() {
  pluginCatalogReadAt=null;pluginCatalogStateAt=null;pluginCatalogReadFailed=false;pluginCatalogLoading=null;
  renderPluginCatalogControls();
}
function pluginCatalogReady() {
  return authenticated && !!pluginCatalogReadAt && !pluginCatalogLoading && !pluginSaving && !pluginScopeBusy && !$("plugins-view").hidden && !document.hidden;
}
function renderPluginCatalogControls() {
  $("download-plugin-catalog").disabled=!pluginCatalogReady();
  $("plugin-catalog-status").textContent="";
  $("plugin-catalog-source").textContent=pluginCatalogReadAt ? "清单范围："+(pluginGroup==="0" ? "全局" : "群 "+pluginGroup)+"；最近成功读取："+new Date(pluginCatalogReadAt).toLocaleString("zh-CN",{hour12:false})+"。"+(pluginCatalogStateAt ? "最近成功保存开关："+new Date(pluginCatalogStateAt).toLocaleString("zh-CN",{hour12:false})+"（只更新已保存项）。" : "")+(pluginCatalogReadFailed ? "最近读取失败，清单仍使用当前已加载范围的旧结果。" : "")+"不是服务端实时快照。" : "成功读取插件范围后可下载清单。";
}
function downloadPluginCatalog() {
  if (!pluginCatalogReady()) return;
  let url=null,link=null;
  try {
    const exportedAt=new Date().toISOString(),items=filteredPluginItems();
    const mode=value=>["inherit","enabled","disabled"].includes(value) ? value : null;
    const entries=items.map(plugin=>{
      const known=pluginEffectiveState(plugin)!=="unknown",state=known ? plugin.state : null;
      return {name:typeof plugin.name==="string" ? plugin.name : null,
        defaultEnabledAtStartup:typeof plugin.defaultEnabledAtStartup==="boolean" ? plugin.defaultEnabledAtStartup : null,
        effectiveState:pluginEffectiveState(plugin),defaultEnabled:typeof state?.defaultEnabled==="boolean" ? state.defaultEnabled : null,
        globalMode:mode(state?.globalMode),groupMode:pluginGroup==="0" ? null : mode(state?.groupMode),
        source:["global","group","default"].includes(state?.source) ? state.source : null,
        hasUnsavedChanges:pluginDrafts.has(plugin.name)};
    });
    const value={format:"zerobot-plugin-state-catalog",version:1,exportedAt,readAt:pluginCatalogReadAt,lastSuccessfulSwitchSaveAt:pluginCatalogStateAt,
      scope:{kind:pluginGroup==="0" ? "global" : "group",groupId:pluginGroup==="0" ? null : pluginGroup},
      source:{loaded:plugins.length,matched:entries.length,unsaved:pluginDrafts.size,lastReadFailed:pluginCatalogReadFailed,liveSnapshot:false},
      filters:{state:["all","enabled","disabled","unknown"].includes(pluginStateFilter) ? pluginStateFilter : "all",keywordApplied:!!pluginSearchQuery,onlyUnsaved:pluginOnlyDrafts,sort:["default","name","enabled","disabled","drafts"].includes(pluginSort) ? pluginSort : "default"},plugins:entries};
    // Export an explicit metadata allowlist, never serialize plugin.state or drafts.
    url=URL.createObjectURL(new Blob([JSON.stringify(value,null,2)+"\n"],{type:"application/json;charset=utf-8"}));
    link=document.createElement("a");link.href=url;link.download="zerobot-plugin-state-catalog-"+exportedAt.replace(/[-:.]/g,"")+".json";
    document.body.append(link);link.click();
    $("plugin-catalog-status").textContent="已发起清单下载，共 "+entries.length+" 个匹配插件（跨全部已加载页）。"+(pluginCatalogReadFailed ? "最近读取失败，使用当前已加载范围的旧结果。" : "")+"未保存标记不表示已生效；清单不能导入或恢复。";
  } catch {
    $("plugin-catalog-status").textContent="清单下载未完成，请检查浏览器下载设置后重试。";
  } finally {
    link?.remove();if(url) setTimeout(()=>URL.revokeObjectURL(url),0);
  }
}
function renderPlugins() {
  renderPluginResponse();
  const query = $("plugin-search").value.toLocaleLowerCase().trim();
  const filter = $("plugin-state-filter").value;
  if (query !== pluginSearchQuery || filter !== pluginStateFilter) pluginPage = 1;
  pluginSearchQuery = query; pluginStateFilter = filter;
  const counts = {enabled:0, disabled:0, unknown:0};
  for (const plugin of plugins) counts[pluginEffectiveState(plugin)]++;
  const items = filteredPluginItems();
  $("search-count").textContent = items.length + " / " + plugins.length + " 个插件；启用 " + counts.enabled + "，禁用 " + counts.disabled + "，未知 " + counts.unknown;
  $("plugin-empty").textContent = plugins.length ? "没有符合搜索和状态条件的插件，可调整条件或清除筛选。" : "当前程序未注册插件。";
  if (pluginReloadAfterRestore && !plugins.length && !pluginCatalogReadAt) $("plugin-empty").textContent = pluginScopeBusy || pluginCatalogLoading ? "正在读取插件状态…" : pluginCatalogReadFailed ? "未能读取插件状态，请点击刷新重试。" : "插件状态尚未读取，请点击刷新。";
  if (plugins.length && pluginOnlyDrafts) $("plugin-empty").textContent = pluginDrafts.size ? "没有符合条件的待保存插件，可调整条件或清除筛选；其他草稿仍保留。" : "当前范围没有未保存的开关更改。取消只看未保存可浏览全部插件。";
  $("plugin-list").replaceChildren(); $("plugin-empty").hidden = items.length !== 0;
  pluginMatchCount = items.length; pluginPageCount = Math.max(1, Math.ceil(items.length / pluginPageSize)); pluginPage = Math.min(pluginPage, pluginPageCount);
  renderPluginPagination();
  for (const plugin of items.slice((pluginPage - 1) * pluginPageSize, pluginPage * pluginPageSize)) {
    const card = document.createElement("article"), title = document.createElement("h2"), brief = document.createElement("p"), footer = document.createElement("footer"), badge = document.createElement("span"), button = document.createElement("button");
    card.className = "plugin-card"; card.dataset.plugin = plugin.name; title.textContent = plugin.name; brief.textContent = plugin.brief || "暂无功能简介";
    badge.className = "plugin-default" + (plugin.defaultEnabledAtStartup ? "" : " disabled");
    const state = plugin.state;
    if (state) {
      badge.className = "plugin-default" + (state.effectiveEnabled ? "" : " disabled");
      badge.textContent = state.effectiveEnabled ? "当前生效：启用" : "当前生效：禁用";
      const reason = document.createElement("p"); reason.className = "field-help";
      reason.textContent = "生效来源：" + ({global:"全局覆盖", group:"此群设置", default:"默认值"}[state.source]) + "；当前默认" + (state.defaultEnabled ? "启用" : "禁用");
      if (pluginGroup !== "0" && state.globalMode !== "inherit") reason.textContent += "。全局覆盖中，此群设置暂不决定生效状态。";
      card.append(reason);
    } else badge.textContent = plugin.defaultEnabledAtStartup ? "启动时默认启用" : "启动时默认禁用";
    button.className = "text-button"; button.textContent = "查看用法 →";
    button.addEventListener("click", () => openPluginUsage(plugin));
    footer.append(badge, button); card.prepend(title, brief); card.append(footer);
    if (pluginControlAvailable && state) {
      const controls = document.createElement("div"), select = document.createElement("select"), save = document.createElement("button");
      controls.className = "plugin-controls"; select.setAttribute("aria-label", plugin.name + " 的开关设置");
      for (const [mode, label] of [["inherit", pluginGroup === "0" ? "继承默认" : "继承全局或默认"], ["enabled","显式启用"], ["disabled","显式禁用"]]) {
        const option = document.createElement("option"); option.value = mode; option.textContent = label; select.append(option);
      }
      select.value = pluginDrafts.get(plugin.name)?.mode || (pluginGroup === "0" ? state.globalMode : state.groupMode);
      select.disabled = pluginSaving || pluginScopeBusy; save.disabled = pluginSaving || pluginScopeBusy; save.className = "secondary"; save.textContent = "保存开关";
      select.addEventListener("change", () => {
        const baseline = pluginGroup === "0" ? state.globalMode : state.groupMode;
        if (select.value === baseline) pluginDrafts.delete(plugin.name);
        else pluginDrafts.set(plugin.name, { mode:select.value, revision:pluginDrafts.get(plugin.name)?.revision || state.revision });
        if (pluginOnlyDrafts || pluginSort === "drafts") renderPlugins(); else renderPluginDrafts();
      });
      save.addEventListener("click", () => savePlugin(plugin, select.value));
      controls.append(select,save); card.append(controls);
      const draftActions = document.createElement("div"), note = document.createElement("p"), discard = document.createElement("button");
      draftActions.className = "plugin-single-draft";
      note.className = "field-help plugin-draft-note"; note.textContent = "未保存";
      discard.type = "button"; discard.className = "secondary plugin-discard-one"; discard.textContent = "放弃此项更改";
      discard.dataset.plugin = plugin.name; discard.setAttribute("aria-label", "放弃 " + plugin.name + " 的未保存更改");
      discard.addEventListener("click", () => discardOnePluginDraft(plugin));
      draftActions.append(note,discard); card.append(draftActions);
    }
    $("plugin-list").append(card);
  }
  renderPluginDrafts();
}
function renderPluginDrafts() {
  renderPluginCatalogControls();
  $("plugin-draft-status").textContent = pluginDrafts.size ? "未保存更改：" + pluginDrafts.size + " 个插件（" + (pluginGroup === "0" ? "全局" : "群 " + pluginGroup) + "）。可单项保存或逐项保存全部草稿；搜索、筛选和切换页面会保留选择。" : "没有未保存的开关更改。";
  for (const card of $("plugin-list").children) {
    const note = card.querySelector(".plugin-draft-note"); if (note) note.hidden = !pluginDrafts.has(card.dataset.plugin);
  }
  renderPluginDraftActions();
}
function renderPluginDraftActions() {
  const busy = pluginSaving || pluginScopeBusy;
  for (const button of $("plugin-list").querySelectorAll(".plugin-discard-one")) button.disabled = busy || !authenticated || !pluginControlAvailable || !pluginDrafts.has(button.dataset.plugin);
  $("plugin-only-drafts").disabled = busy || !authenticated;
  $("discard-plugin-drafts").disabled = busy || !authenticated || !pluginDrafts.size;
  $("save-all-plugin-drafts").disabled = busy || !authenticated || !pluginControlAvailable || !pluginDrafts.size;
  $("stop-plugin-batch").disabled = !authenticated || !pluginBatch || pluginBatch.stop;
}
function pluginBatchSnapshot() { return JSON.stringify([...pluginDrafts]); }
function pluginBatchCurrent(batch) {
  return pluginBatch === batch && authenticated && csrf === batch.token && pluginGeneration === batch.generation && pluginGroup === batch.group;
}
function pluginBatchResultValid(value, item, group) {
  const state = value?.state;
  return value?.name === item.name && value?.groupId === group && state &&
    typeof state.effectiveEnabled === "boolean" && /^[a-f0-9]{64}$/.test(state.revision || "") &&
    ["global", "group", "default"].includes(state.source) &&
    (group === "0" ? state.globalMode : state.groupMode) === item.mode;
}
async function saveAllPluginDrafts() {
  if (!authenticated || pluginSaving || pluginScopeBusy || !pluginControlAvailable || !pluginDrafts.size) return;
  const generation = pluginGeneration, group = pluginGroup, token = csrf, snapshot = pluginBatchSnapshot(), list = plugins;
  const items = plugins.filter(p => pluginDrafts.has(p.name)).map(p => ({name:p.name, plugin:p, ...pluginDrafts.get(p.name)}));
  if (items.length !== pluginDrafts.size || new Set(items.map(p => p.name)).size !== items.length || items.some(p => !p.plugin.state || !["inherit", "enabled", "disabled"].includes(p.mode) || !/^[a-f0-9]{64}$/.test(p.revision || ""))) {
    showError("plugin-error", new Error("草稿与已读取插件状态不一致，请先核对并刷新状态。")); return;
  }
  const scope = group === "0" ? "全局（覆盖所有群与私聊）" : "群 " + group;
  if (!window.confirm("逐项保存“" + scope + "”的全部 " + items.length + " 个插件草稿？包含其他页和筛选隐藏项，每项分别备份并立即生效；遇错停止，已成功规则不回滚。")) return;
  if (!authenticated || pluginSaving || pluginScopeBusy || token !== csrf || generation !== pluginGeneration || group !== pluginGroup || list !== plugins || snapshot !== pluginBatchSnapshot()) return;
  const batch = {generation, group, token, items, completed:0, stop:false};
  pluginBatch = batch; pluginSaving = true; pluginListRequest++; $("plugin-error").hidden = true;
  renderPluginBusy(); renderPlugins();
  try {
    for (const item of items) {
      if (!pluginBatchCurrent(batch)) return;
      if (batch.stop) break;
      const draft = pluginDrafts.get(item.name);
      if (!draft || draft.mode !== item.mode || draft.revision !== item.revision) throw new Error("待保存草稿已变化，后续操作已停止，请核对余下选择。");
      $("plugin-batch-status").textContent = scope + "：已确认成功 " + batch.completed + " / " + items.length + " 项，正在保存 " + item.name + "。";
      const controller = new AbortController(), timer = setTimeout(() => controller.abort(), 30000);
      let value;
      try { value = await api("/api/plugin-switch", {name:item.name, groupId:group, mode:item.mode, revision:item.revision}, controller.signal); }
      finally { clearTimeout(timer); }
      if (!pluginBatchCurrent(batch)) return;
      if (!pluginBatchResultValid(value, item, group)) throw new Error("返回结果无法核对，请先重新读取状态，勿直接重复保存。");
      item.plugin.state = value.state;
      pluginCatalogStateAt=new Date().toISOString();
      pluginDrafts.delete(item.name); batch.completed++; renderPlugins();
    }
    if (!pluginBatchCurrent(batch)) return;
    $("plugin-batch-status").textContent = batch.completed === items.length ? scope + "：全部 " + items.length + " 项已确认保存，规则立即更新。" : scope + "：已停止后续保存；已确认成功 " + batch.completed + " / " + items.length + " 项，余下 " + (items.length - batch.completed) + " 项草稿保留，已成功规则不回滚。";
  } catch (error) {
    if (pluginBatchCurrent(batch)) {
      $("plugin-batch-status").textContent = scope + "：保存已停止；已确认成功 " + batch.completed + " / " + items.length + " 项，余下 " + (items.length - batch.completed) + " 项未确认、草稿保留。已提交项可能已生效，请核对状态；不会自动重试或回滚。";
      showError("plugin-error", error.name === "AbortError" ? new Error("当前项请求超时或中断，结果未确认，请重新读取状态核对。") : error);
    }
  } finally {
    if (pluginBatchCurrent(batch)) { pluginBatch = null; pluginSaving = false; renderPluginBusy(); renderPlugins(); }
  }
}
function stopPluginBatch() {
  if (!pluginBatch || !pluginBatchCurrent(pluginBatch) || pluginBatch.stop) return;
  pluginBatch.stop = true;
  $("plugin-batch-status").textContent = "已请求停止后续保存；等待当前请求结束。当前项可能生效，已成功规则不会撤销。";
  renderPluginDraftActions();
}
function discardOnePluginDraft(plugin) {
  if (!authenticated || pluginSaving || pluginScopeBusy || !pluginControlAvailable || !plugin?.state || !plugins.includes(plugin) || !pluginDrafts.has(plugin.name)) return;
  const generation = pluginGeneration, group = pluginGroup, token = csrf, list = plugins, name = plugin.name;
  const snapshot = JSON.stringify({draft:pluginDrafts.get(name),state:plugin.state});
  const scope = group === "0" ? "全局" : "群 " + group;
  if (!window.confirm("放弃「" + name + "」（" + scope + "）的未保存更改？仅恢复页面最近读取或保存的选择，其他插件草稿保留。")) return;
  if (!authenticated || pluginSaving || pluginScopeBusy || !pluginControlAvailable || generation !== pluginGeneration || group !== pluginGroup || token !== csrf || list !== plugins || name !== plugin.name || !plugins.includes(plugin) || snapshot !== JSON.stringify({draft:pluginDrafts.get(name),state:plugin.state})) return;
  pluginDrafts.delete(name); renderPlugins();
  $("plugin-save-state").textContent = "已放弃「" + name + "」（" + scope + "）的未保存选择；其他草稿保留，已保存规则未改变。";
}
function discardPluginDrafts() {
  if (!authenticated || pluginSaving || pluginScopeBusy || !pluginDrafts.size) return;
  const generation = pluginGeneration, group = pluginGroup;
  if (!confirmPluginDiscard("放弃全部更改")) return;
  if (!authenticated || pluginSaving || pluginScopeBusy || generation !== pluginGeneration || group !== pluginGroup) return;
  const count = pluginDrafts.size;
  pluginDrafts.clear(); renderPlugins();
  $("plugin-save-state").textContent = "已放弃 " + count + " 个插件的未保存选择；已保存的开关规则未改变。";
}
function renderPluginPagination() {
  const busy = pluginSaving || pluginScopeBusy;
  $("plugin-page-status").textContent = pluginMatchCount ? "第 " + pluginPage + " / " + pluginPageCount + " 页；显示 " + ((pluginPage - 1) * pluginPageSize + 1) + "～" + Math.min(pluginPage * pluginPageSize, pluginMatchCount) + " / " + pluginMatchCount + " 个匹配插件。" : "没有匹配插件。";
  $("plugin-prev").disabled = busy || !pluginMatchCount || pluginPage <= 1;
  $("plugin-next").disabled = busy || !pluginMatchCount || pluginPage >= pluginPageCount;
  $("plugin-page-size").disabled = busy;
  $("plugin-sort").disabled = busy || !authenticated;
  $("plugin-sort").value = pluginSort;
}
function changePluginPage(step) {
  if (!authenticated || pluginSaving || pluginScopeBusy) return;
  const next = pluginPage + step;
  if (next < 1 || next > pluginPageCount) return;
  pluginPage = next; renderPlugins();
}
function renderPluginScope() {
  $("plugin-scope").value = pluginGroup === "0" ? "global" : "group";
  if (pluginGroup !== "0") $("plugin-group").value = pluginGroup;
  $("plugin-scope-state").textContent = pluginGroup === "0" ? "当前范围：全局。显式全局设置会覆盖所有群与私聊的插件开关。" : "当前范围：群 " + pluginGroup + "。显式全局设置优先生效。";
}
function renderPluginResponse() {
  $("plugin-response-panel").hidden = !authenticated || !pluginResponseLoaded;
  $("plugin-response-state").textContent = ""; $("plugin-response-detail").textContent = "";
  $("plugin-response-title").textContent = "响应状态说明";
  if (!authenticated || !pluginResponseLoaded) return;
  const scope = pluginGroup === "0" ? "全局" : "群 " + pluginGroup;
  $("plugin-response-title").textContent = "响应状态 · " + scope;
  const value = pluginResponse;
  const valid = value && [value.globalEnabled, value.groupEnabled, value.effectiveEnabled].every(v => typeof v === "boolean") &&
    value.effectiveEnabled === (value.globalEnabled || value.groupEnabled) &&
    value.source === (value.globalEnabled ? "global" : value.groupEnabled ? "group" : "none") &&
    (pluginGroup !== "0" || !value.groupEnabled);
  if (!valid) { $("plugin-response-state").textContent = "当前实例未提供可核对的统一响应状态，请使用现有 QQ 指令或日志核对。"; return; }
  if (pluginGroup === "0") {
    $("plugin-response-state").textContent = "全局响应：" + (value.globalEnabled ? "开启" : "关闭") + "。";
    $("plugin-response-detail").textContent = value.globalEnabled ? "已开启全局响应；各群单独响应关闭也不能阻止全局响应。" : "全局入口未开启；部分群可能仍单独响应，请读取具体群核对。";
  } else {
    $("plugin-response-state").textContent = "全局响应：" + (value.globalEnabled ? "开启" : "关闭") + "；此群单独响应：" + (value.groupEnabled ? "开启" : "关闭") + "；此群响应检查：" + (value.effectiveEnabled ? "通过" : "未通过") + "。";
    $("plugin-response-detail").textContent = value.source === "global" ? "响应来源：全局。此群单独关闭不能覆盖已开启的全局响应。" : value.source === "group" ? "响应来源：此群单独设置。全局关闭不清除此群已开启的响应。" : "当前未开启全局或此群响应；仅启用插件不会自动打开响应。";
  }
}
function renderPluginBusy() {
  const busy = pluginSaving || pluginScopeBusy;
  $("apply-plugin-scope").disabled = busy; $("reload-plugins").disabled = busy;
  $("plugin-scope").disabled = busy; $("plugin-group").disabled = busy || $("plugin-scope").value !== "group";
  $("plugin-group").required = $("plugin-scope").value === "group";
  renderPluginPagination();
  renderPluginDraftActions();
}
function confirmPluginDiscard(action) {
  if (pluginSaving || pluginScopeBusy) { $("plugin-save-state").textContent = "插件请求尚未完成，请稍候再" + action + "。"; return false; }
  return !pluginDrafts.size || window.confirm("有 " + pluginDrafts.size + " 个插件开关未保存。" + action + "将丢弃这些选择，是否继续？");
}
function pluginEffectiveState(plugin) {
  return typeof plugin.state?.effectiveEnabled === "boolean" ? (plugin.state.effectiveEnabled ? "enabled" : "disabled") : "unknown";
}
async function loadPluginList(group = pluginGroup, discard = false) {
  const request = ++pluginListRequest, generation = pluginGeneration;
  pluginCatalogLoading=request;renderPluginCatalogControls();
  let list;
  try {list = await api("/api/plugins?group=" + encodeURIComponent(group));}
  catch(error) {
    if (request===pluginListRequest && generation===pluginGeneration && authenticated) {pluginCatalogReadFailed=true;renderPluginCatalogControls();}
    throw error;
  } finally {
    if (pluginCatalogLoading===request && generation===pluginGeneration) {pluginCatalogLoading=null;renderPluginCatalogControls();}
  }
  if (request !== pluginListRequest || generation !== pluginGeneration || pluginSaving || !authenticated) return false;
  if (discard) pluginDrafts.clear();
  if (group !== pluginGroup) pluginPage = 1;
  pluginGroup = group; plugins = list.plugins || []; pluginControlAvailable = !!list.controlAvailable;
  pluginReloadAfterRestore = false;
  pluginCatalogReadAt=new Date().toISOString();pluginCatalogStateAt=null;pluginCatalogReadFailed=false;
  pluginResponse = list.response || null; pluginResponseLoaded = true;
  $("plugin-batch-status").textContent = "";
  renderPluginScope(); renderPlugins(); return true;
}
async function ensurePluginList() {
  if (!authenticated || !pluginReloadAfterRestore || pluginCatalogLoading || pluginSaving || pluginScopeBusy) return;
  const generation = pluginGeneration, group = pluginGroup;
  pluginScopeBusy = true; $("plugin-error").hidden = true;
  renderPluginBusy(); renderPlugins();
  try { await loadPluginList(group, true); }
  catch (error) { if (generation === pluginGeneration && authenticated) showError("plugin-error", error); }
  finally { if (generation === pluginGeneration) { pluginScopeBusy = false; renderPluginBusy(); renderPlugins(); } }
}
async function savePlugin(plugin, mode) {
  if (pluginSaving || pluginScopeBusy || !authenticated) return;
  const revision = pluginDrafts.get(plugin.name)?.revision || plugin.state.revision, generation = pluginGeneration, group = pluginGroup;
  pluginSaving = true; pluginListRequest++; $("plugin-error").hidden = true; $("plugin-save-state").textContent = "正在保存 " + plugin.name + "…";
  renderPluginBusy(); renderPlugins();
  try {
    const value = await api("/api/plugin-switch", {name:plugin.name,groupId:group,mode,revision});
    if (generation !== pluginGeneration || !authenticated) return;
    plugin.state = value.state; pluginDrafts.delete(plugin.name);pluginCatalogStateAt=new Date().toISOString();
    $("plugin-save-state").textContent = plugin.name + " 已保存，运行规则已更新。" + (pluginGroup !== "0" && value.state.source === "global" ? "当前仍由全局设置决定生效状态。" : "");
  } catch (error) { if (generation === pluginGeneration && authenticated) { $("plugin-save-state").textContent = "保存未完成。"; showError("plugin-error", error); } }
  finally { if (generation === pluginGeneration) { pluginSaving = false; renderPluginBusy(); renderPlugins(); } }
}
$("plugin-scope").addEventListener("change", renderPluginBusy);
$("plugin-scope-form").addEventListener("submit", async (event) => {
  event.preventDefault(); if (pluginSaving || pluginScopeBusy || !authenticated) return;
  const group = $("plugin-scope").value === "global" ? "0" : $("plugin-group").value.trim();
  if ($("plugin-scope").value === "group" && !/^[1-9][0-9]{0,18}$/.test(group)) { showError("plugin-error",new Error("请填写有效的群号。")); return; }
  if (!confirmPluginDiscard("读取插件范围")) { renderPluginScope(); renderPluginBusy(); return; }
  $("plugin-error").hidden = true;
  const generation = pluginGeneration;
  pluginScopeBusy = true; renderPluginBusy(); renderPlugins();
  try {
    if (!await loadPluginList(group, true)) return;
    $("plugin-save-state").textContent = "";
  } catch (error) { if (generation === pluginGeneration && authenticated) { renderPluginScope(); showError("plugin-error",error); } }
  finally { if (generation === pluginGeneration) { pluginScopeBusy = false; renderPluginBusy(); renderPlugins(); } }
});
$("reload-plugins").addEventListener("click", async () => {
  if (!authenticated || !confirmPluginDiscard("刷新插件状态")) return;
  const generation = pluginGeneration;
  pluginScopeBusy = true; renderPluginBusy(); renderPlugins();
  $("plugin-error").hidden = true;
  try { if (await loadPluginList(pluginGroup, true)) $("plugin-save-state").textContent="状态已重新加载。"; }
  catch (error) { if (generation === pluginGeneration && authenticated) showError("plugin-error",error); }
  finally { if (generation === pluginGeneration) { pluginScopeBusy = false; renderPluginBusy(); renderPlugins(); } }
});
window.addEventListener("beforeunload", event => {
  if (authenticated && (pluginSaving || pluginScopeBusy || pluginDrafts.size || connectionBusy || connectionDirty())) { event.preventDefault(); event.returnValue = ""; }
});
$("auth-form").addEventListener("submit", async (event) => {
  event.preventDefault(); $("auth-error").hidden = true;
  if (setupRequired && $("password").value !== $("confirm-password").value) { showError("auth-error", new Error("两次输入的密码不一致。")); return; }
  $("login-button").disabled = true;
  try {
    const value = await api(setupRequired ? "/api/setup" : "/api/login", setupRequired ? { setupCode: $("setup-code").value, password: $("password").value } : { password: $("password").value });
    csrf = value.csrfToken; await showWorkspace();
  } catch (error) { showError("auth-error", error); }
  finally { $("login-button").disabled = false; }
});
$("setup-file").addEventListener("change", async () => {
  const file = $("setup-file").files[0];
  if (!file) return;
  if (file.size > 1024) { showError("auth-error", new Error("请选择后台生成的设置码文件。")); return; }
  $("setup-code").value = (await file.text()).trim();
});
$("logout").addEventListener("click", async () => {
  if (typeof securityCanLeave === "function" && !securityCanLeave()) return;
  const session = csrf;
  if (!await confirmConnectionDiscard("退出登录")) return;
  if (!confirmPluginDiscard("退出登录")) return;
  if (typeof confirmAIDiscard === "function" && !confirmAIDiscard("退出登录")) return;
  if (!authenticated || session !== csrf) return;
  try { await api("/api/logout", {}); if (!authenticated || session !== csrf) return; authenticated = false; csrf = ""; showAuth(); }
  catch (error) { if (authenticated && session === csrf) showError("page-error", error); }
});
$("refresh").addEventListener("click", () => refreshOverview());
$("plugin-search").addEventListener("input", renderPlugins);
$("plugin-sort").addEventListener("change", () => {
  const order = $("plugin-sort").value;
  if (!authenticated || pluginSaving || pluginScopeBusy || !["default","name","enabled","disabled","drafts"].includes(order)) { $("plugin-sort").value = pluginSort; return; }
  pluginSort = order; pluginPage = 1; renderPlugins();
});
$("plugin-state-filter").addEventListener("change", renderPlugins);
$("clear-plugin-filter").addEventListener("click", () => {
  pluginOnlyDrafts = false; $("plugin-only-drafts").checked = false;
  pluginPage = 1; $("plugin-search").value = ""; $("plugin-state-filter").value = "all"; renderPlugins();
});
$("plugin-only-drafts").addEventListener("change", () => {
  if (!authenticated || pluginSaving || pluginScopeBusy) { $("plugin-only-drafts").checked = pluginOnlyDrafts; return; }
  pluginOnlyDrafts = $("plugin-only-drafts").checked; pluginPage = 1; renderPlugins();
});
$("discard-plugin-drafts").addEventListener("click", discardPluginDrafts);
$("save-all-plugin-drafts").addEventListener("click", saveAllPluginDrafts);
$("stop-plugin-batch").addEventListener("click", stopPluginBatch);
$("plugin-prev").addEventListener("click", () => changePluginPage(-1));
$("plugin-next").addEventListener("click", () => changePluginPage(1));
$("plugin-page-size").addEventListener("change", () => {
  if (!authenticated || pluginSaving || pluginScopeBusy) return;
  const size = Number($("plugin-page-size").value);
  if (![10, 20, 50].includes(size)) { $("plugin-page-size").value = String(pluginPageSize); return; }
  pluginPageSize = size; pluginPage = 1; renderPlugins();
});
function resetPluginUsage() {
  pluginUsage = null;
  $("usage-title").textContent = ""; $("usage-text").textContent = "";
  $("usage-export-status").textContent = "";
  $("copy-plugin-usage").disabled = true; $("download-plugin-usage").disabled = true;
  const selection = window.getSelection();
  if (selection?.anchorNode && $("usage-text").contains(selection.anchorNode)) selection.removeAllRanges();
}
function closePluginUsage() {
  resetPluginUsage();
  if ($("usage-dialog").open) $("usage-dialog").close();
}
function openPluginUsage(plugin) {
  if (!authenticated) return;
  const name = typeof plugin.name === "string" ? plugin.name : "插件";
  const brief = typeof plugin.brief === "string" ? plugin.brief : "";
  const usage = typeof plugin.usage === "string" ? plugin.usage : "";
  const safeName = Array.from(name.replace(/[<>:"/\\|?*\u0000-\u001f\u007f]/g, "_")).slice(0, 64).join("").replace(/[. ]+$/g, "") || "usage";
  pluginUsage = {name, usage, token:csrf, text:`插件：${name}\n简介：${brief}\n\n注册用法：\n${usage}\n`, filename:`zerobot-plugin-${safeName}.txt`, copying:false};
  $("usage-title").textContent = name;
  $("usage-text").textContent = usage.trim() ? usage : "这个插件暂未提供用法说明。";
  $("usage-export-status").textContent = "";
  $("copy-plugin-usage").disabled = !usage.trim(); $("download-plugin-usage").disabled = !usage.trim();
  if (!$("usage-dialog").open) $("usage-dialog").showModal();
}
function pluginUsageCurrent(snapshot) {
  return authenticated && pluginUsage === snapshot && csrf === snapshot.token && $("usage-dialog").open;
}
async function copyPluginUsage() {
  const snapshot = pluginUsage;
  if (!snapshot || !snapshot.usage.trim() || snapshot.copying || !pluginUsageCurrent(snapshot)) return;
  snapshot.copying = true; $("copy-plugin-usage").disabled = true;
  $("usage-export-status").textContent = "正在复制用法…";
  try {
    if (!navigator.clipboard?.writeText) throw new Error("clipboard unavailable");
    await navigator.clipboard.writeText(snapshot.usage);
    if (pluginUsageCurrent(snapshot)) $("usage-export-status").textContent = "已复制注册用法。";
  } catch {
    if (pluginUsageCurrent(snapshot)) {
      const range = document.createRange(); range.selectNodeContents($("usage-text"));
      const selection = window.getSelection(); selection?.removeAllRanges(); selection?.addRange(range);
      $("usage-text").focus();
      $("usage-export-status").textContent = "复制未完成，已选中用法；请按 Ctrl+C 手动复制，或下载文本。";
    }
  } finally {
    if (pluginUsageCurrent(snapshot)) { snapshot.copying = false; $("copy-plugin-usage").disabled = false; }
  }
}
function downloadPluginUsage() {
  const snapshot = pluginUsage;
  if (!snapshot || !snapshot.usage.trim() || !pluginUsageCurrent(snapshot)) return;
  let url = "", link;
  try {
    url = URL.createObjectURL(new Blob([snapshot.text], {type:"text/plain;charset=utf-8"}));
    link = document.createElement("a"); link.href = url; link.download = snapshot.filename; link.hidden = true;
    document.body.append(link); link.click();
    $("usage-export-status").textContent = "已发起文本下载，请在浏览器下载列表核对。";
  } catch {
    $("usage-export-status").textContent = "下载未完成，请重试或手动复制用法。";
  } finally {
    link?.remove();
    if (url) setTimeout(() => URL.revokeObjectURL(url), 1000);
  }
}
$("close-dialog").addEventListener("click", closePluginUsage);
$("usage-dialog").addEventListener("close", () => { if (!$("usage-dialog").open) resetPluginUsage(); });
$("copy-plugin-usage").addEventListener("click", copyPluginUsage);
$("download-plugin-usage").addEventListener("click", downloadPluginUsage);
$("download-plugin-catalog").addEventListener("click", downloadPluginCatalog);
document.addEventListener("visibilitychange",renderPluginCatalogControls);
document.querySelectorAll("[data-view]").forEach((button) => button.addEventListener("click", () => {
  if (typeof connectionDeleteCanLeave === "function" && !connectionDeleteCanLeave()) return;
  if (typeof connectionAddCanLeave === "function" && !connectionAddCanLeave()) return;
  if (typeof securityCanLeave === "function" && !securityCanLeave()) return;
  document.querySelectorAll("[data-view]").forEach((item) => item.classList.toggle("active", item === button));
  for (const view of ["overview", "plugins", "connections", "ai", "logs", "backups", "security"]) $(view + "-view").hidden = button.dataset.view !== view;
  renderPluginCatalogControls();
  $("breadcrumb").textContent = "工作空间 / " + ({ overview:"运行概览", plugins:"插件列表", connections:"连接设置", ai:"AI 设置", logs:"运行日志", backups:"备份恢复", security:"后台安全" }[button.dataset.view]);
  showOverview(button.dataset.view === "overview");
  if (button.dataset.view === "plugins") ensurePluginList();
  if (button.dataset.view === "connections" && !connectionSettings) loadConnections();
  if (button.dataset.view === "ai" && !aiState) loadAI();
  if (typeof showLogs === "function") showLogs(button.dataset.view === "logs");
  if (typeof showBackups === "function") showBackups(button.dataset.view === "backups");
  if (typeof showSecurity === "function") showSecurity(button.dataset.view === "security");
  if (typeof showConnectionDelete === "function") showConnectionDelete(button.dataset.view === "connections");
}));

function selectedConnection() { return connectionSettings?.connections?.[connectionIndex]; }
function connectionSnapshot() {
  return JSON.stringify([$("connection-url").value, $("token-action").value, $("connection-token").value, $("clear-token-confirm").checked]);
}
function connectionDirty(connectionOnly = false) {
  if (!connectionSettings?.available) return false;
  const changed = connectionBaseline !== null && connectionSnapshot() !== connectionBaseline;
  return changed || (!connectionOnly && (connectionOwnersDirty() || (typeof connectionAddDirty === "function" && connectionAddDirty())));
}
function connectionOwnersDirty() {
  return !!connectionSettings?.available && $("connection-owners").value !== (connectionSettings.superUsers || []).join(", ");
}
function renderConnectionDrafts() {
  const parts = [];
  if (connectionDirty(true)) parts.push("当前连接地址或 Token");
  if (connectionOwnersDirty()) parts.push("主人账号");
  if (typeof connectionAddDirty === "function" && connectionAddDirty()) parts.push("新增连接");
  $("connection-draft-status").textContent = parts.length ? "未保存更改：" + parts.join("、") + "。保存前请核对；切换后台页面保留草稿。" : "没有未保存的连接配置更改。";
  renderConnectionDraftFields();
  renderConnectionDiscardControls();
  if (typeof renderConnectionAddControls === "function") renderConnectionAddControls();
  if (typeof renderConnectionDeleteControls === "function") renderConnectionDeleteControls();
}
function renderConnectionDiscardControls() {
  const blocked = !authenticated || connectionBusy || !connectionSettings?.available;
  $("discard-current-connection").disabled = blocked || !selectedConnection() || !connectionDirty(true);
  $("discard-connection-owners").disabled = blocked || !connectionOwnersDirty();
}
function renderConnectionDraftFields() {
  const list = $("connection-draft-field-list"); list.replaceChildren();
  $("connection-draft-fields").hidden = true;
  $("connection-draft-field-state").textContent = "";
  if (!authenticated || !connectionSettings?.available) return;
  const entries = [];
  if (selectedConnection() && connectionBaseline !== null && connectionDirty(true)) {
    const before = JSON.parse(connectionBaseline), current = JSON.parse(connectionSnapshot()), fields = [];
    if (before[0] !== current[0]) fields.push("WebSocket 地址");
    if ([1, 2, 3].some(index => before[index] !== current[index])) fields.push("Token");
    const connection = selectedConnection();
    const known = ["ws", "wss"].includes(connection.kind) && Number.isSafeInteger(connection.index) && connection.index >= 0 && connection.index < Number.MAX_SAFE_INTEGER;
    const title = known ? (connection.kind === "ws" ? "正向连接 " : "反向监听 ") + (connection.index + 1) : "当前连接（目标暂不可核对）";
    entries.push({part:"connection", title, fields});
  }
  if (connectionOwnersDirty()) entries.push({part:"owners", title:"主人账号 · 启动配置", fields:["主人 QQ 账号"]});
  if (typeof connectionAddDirty === "function" && connectionAddDirty()) {
    const fields = [];
    if ($("connection-add-kind").value !== "ws") fields.push("连接方向");
    if ($("connection-add-url").value) fields.push("WebSocket 地址");
    if ($("connection-add-token").value || $("connection-add-without-token").checked) fields.push("Token");
    entries.push({part:"addition", title:"尚未保存的新连接", fields});
  }
  if (!entries.length) return;
  $("connection-draft-fields").hidden = false;
  $("connection-draft-field-state").textContent = connectionBusy ? "请求处理中，保留当前字段摘要，请等待结果；不会自动保存。" : "仅比较当前输入与页面基线；保存后仍按原提示重启生效。";
  for (const entry of entries) {
    const item = document.createElement("article"); item.dataset.connectionDraftPart = entry.part;
    const heading = document.createElement("h3"); heading.textContent = entry.title;
    const count = document.createElement("p"); count.className = "field-help"; count.textContent = "未保存字段：" + entry.fields.length + " 项（Token 操作、输入与确认合并计数）。";
    const fields = document.createElement("ul");
    for (const label of entry.fields) { const field = document.createElement("li"); field.textContent = label; fields.append(field); }
    item.append(heading, count, fields); list.append(item);
  }
}
function finishConnectionDraftConfirmation(accepted = false) {
  const pending = connectionDraftConfirmation;
  connectionDraftConfirmation = null;
  const dialog = $("connection-draft-confirmation");
  if (dialog.open) dialog.close();
  $("connection-draft-confirm-message").textContent = "";
  if (!pending) return;
  pending.resolve(accepted);
  if (authenticated && pending.origin?.isConnected && !pending.origin.disabled) pending.origin.focus();
}
function confirmConnectionDraft(message, label) {
  if (connectionDraftConfirmation || (typeof connectionAddConfirmation !== "undefined" && connectionAddConfirmation) || !authenticated || connectionBusy) return Promise.resolve(false);
  return new Promise(resolve => {
    connectionDraftConfirmation = {resolve, origin:document.activeElement};
    $("connection-draft-confirm-message").textContent = message;
    $("connection-draft-confirm-accept").textContent = label;
    try { $("connection-draft-confirmation").showModal(); }
    catch (_) { finishConnectionDraftConfirmation(); showError("connection-error", new Error("无法打开确认窗口，请重新加载页面后重试。")); }
  });
}
$("connection-draft-confirm-cancel").addEventListener("click", () => finishConnectionDraftConfirmation());
$("connection-draft-confirm-accept").addEventListener("click", () => finishConnectionDraftConfirmation(true));
$("connection-draft-confirmation").addEventListener("cancel", event => { event.preventDefault(); finishConnectionDraftConfirmation(); });
$("connection-draft-confirmation").addEventListener("close", () => { if (!$("connection-draft-confirmation").open) finishConnectionDraftConfirmation(); });
function connectionDraftContext(snapshot) {
  const session = csrf, generation = connectionGeneration, settings = connectionSettings, index = connectionIndex, draft = snapshot();
  return () => authenticated && !connectionBusy && session === csrf && generation === connectionGeneration && settings === connectionSettings && index === connectionIndex && draft === snapshot();
}
async function discardConnectionDraft(part) {
  if (!authenticated || connectionBusy || !connectionSettings?.available) return;
  if (part === "connection" ? !selectedConnection() || !connectionDirty(true) : part !== "owners" || !connectionOwnersDirty()) return;
  const settings = connectionSettings;
  const current = connectionDraftContext(part === "connection" ? connectionSnapshot : () => $("connection-owners").value);
  const connection = selectedConnection();
  const title = part === "owners" ? "主人账号" : (connection.kind === "ws" ? "正向连接 " : "反向监听 ") + (connection.index + 1) + " 的地址与 Token";
  if (!await confirmConnectionDraft("放弃“" + title + "”的未保存更改？另一部分草稿和已保存配置保持不变。", "确认放弃") || !current()) return;
  if (part === "connection") renderSelectedConnection(); else $("connection-owners").value = (settings.superUsers || []).join(", ");
  renderConnectionDrafts();
  $("connection-discard-status").textContent = "已放弃“" + title + "”的未保存更改；已保存配置和运行状态保持不变。";
}
function connectionControls(busy) {
  connectionBusy = busy;
  $("connection-fields").disabled = busy || !connectionSettings?.available;
  $("save-connections").disabled = busy || !connectionSettings?.available;
  $("reload-connections").disabled = busy;
  renderConnectionDraftFields();
  renderConnectionDiscardControls();
  if (typeof renderConnectionAddControls === "function") renderConnectionAddControls();
  if (typeof renderConnectionDeleteControls === "function") renderConnectionDeleteControls();
}
async function confirmConnectionDiscard(action, connectionOnly = false) {
  if (connectionBusy) { $("connection-save-state").textContent = "连接请求尚未完成，请稍候再" + action + "。"; return false; }
  if (connectionDraftConfirmation || (typeof connectionAddConfirmation !== "undefined" && connectionAddConfirmation)) return false;
  const current = connectionDraftContext(connectionOnly ? connectionSnapshot : () => JSON.stringify([connectionSnapshot(), $("connection-owners").value, typeof connectionAddSnapshot === "function" ? connectionAddSnapshot() : null]));
  const accepted = !connectionDirty(connectionOnly) || await confirmConnectionDraft("连接设置有未保存更改。" + action + "将丢弃" + (connectionOnly ? "当前连接的地址与 Token 选择（主人账号草稿保留）" : "这些修改") + "，是否继续？", "确认" + action);
  return accepted && current();
}
function resetConnections() {
  finishConnectionDraftConfirmation();
  connectionGeneration++; connectionSettings = null; connectionIndex = 0; connectionBaseline = null;
  renderStartupDifferences(null);
  if (typeof resetConnectionDelete === "function") resetConnectionDelete();
  if (typeof resetConnectionAdd === "function") resetConnectionAdd();
  $("connection-select").replaceChildren();
  for (const id of ["connection-url", "connection-token", "connection-owners"]) $(id).value = "";
  $("token-action").value = "keep"; updateTokenAction();
  for (const id of ["connection-running", "connection-running-owners", "connection-address-note", "token-state"]) $(id).textContent = "";
  $("connection-save-state").textContent = "请重新读取配置。"; $("connection-error").hidden = true;
  $("connection-discard-status").textContent = "";
  connectionControls(false); renderConnectionDrafts();
}
function updateTokenAction() {
  const mode = $("token-action").value;
  $("connection-token").disabled = mode !== "replace"; $("connection-token").required = mode === "replace";
  if (mode !== "replace") $("connection-token").value = "";
  $("clear-token-label").hidden = mode !== "clear"; $("clear-token-confirm").checked = false;
}
function renderSelectedConnection() {
  const connection = selectedConnection();
  $("connection-select").value = String(connectionIndex);
  $("connection-url").value = connection?.url || ""; $("connection-url").disabled = !connection;
  $("token-action").value = "keep"; $("token-action").disabled = !connection;
  $("connection-token").value = ""; updateTokenAction();
  $("token-state").textContent = connection?.tokenConfigured ? "已配置；当前值不会显示。保留时无需填写。" : "未配置 Token。";
  $("connection-address-note").textContent = connection?.addressRedacted ? "原地址中的认证参数已隐藏。不改动地址时会完整保留；替换地址请使用单独的 Token。" : "支持 ws:// 或 wss://，认证信息填写在 Token 中。";
  const running = connectionSettings?.runningConnections?.find(item => item.kind === connection?.kind && item.index === connection?.index);
  $("connection-running").textContent = running ? "本次启动原序号 " + (connection.kind === "ws" ? "正向连接 " : "反向监听 ") + (connection.index + 1) + " 的地址：" + running.url + "；Token " + (running.tokenConfigured ? "已配置" : "未配置") : "本次启动没有这个连接。";
  connectionBaseline = connectionSnapshot(); renderConnectionDrafts();
}
function renderStartupDifferences(value) {
  const list = $("startup-difference-list"); list.replaceChildren();
  $("startup-differences").hidden = !value?.available;
  $("startup-difference-status").textContent = "";
  if (!value?.available) return;
  const items = value.startupDifferences;
  if (!Array.isArray(items) || !items.length) {
    $("startup-difference-status").textContent = "暂无分类摘要；请以配置保存状态为准。"; return;
  }
  const changed = items.filter(item => item.changed).length;
  $("startup-difference-status").textContent = value.restartRequired ? "已保存配置与本次启动不同，待重启；" + changed + " 个类别有变化。" : "已保存配置与本次启动一致，无需为这些配置重启。";
  for (const item of items) {
    const row = document.createElement("article"); row.className = "startup-difference-row";
    const title = document.createElement("strong"); title.textContent = item.label;
    const state = document.createElement("span"); state.className = "startup-difference-state"; state.classList.toggle("difference-changed", !!item.changed); state.textContent = item.changed ? "有变化" : "一致";
    row.append(title, state);
    if (Number.isSafeInteger(item.runningCount) && item.runningCount >= 0 && Number.isSafeInteger(item.savedCount) && item.savedCount >= 0) {
      const detail = document.createElement("p"); detail.className = "field-help";
      detail.textContent = "本次启动 " + item.runningCount + (item.unit || "") + " → 已保存 " + item.savedCount + (item.unit || "") + (item.changed && item.runningCount === item.savedCount ? "（数量相同，内容不同）" : ""); row.append(detail);
    }
    list.append(row);
  }
}
function renderConnectionSettings(value) {
  connectionSettings = value;
  renderStartupDifferences(value);
  $("connection-discard-status").textContent = "";
  if (connectionIndex >= (value.connections || []).length) connectionIndex = 0;
  $("connection-save-state").textContent = !value.available ? "当前程序未使用启动配置文件，请以 -c 配置文件 启动后再管理连接。" : value.restartRequired ? "配置已保存，与本次启动配置不同，待重启机器人生效。" : "已保存配置与本次启动配置一致。";
  $("connection-select").replaceChildren();
  for (const [index, connection] of (value.connections || []).entries()) {
    const option = document.createElement("option"); option.value = String(index);
    option.textContent = (connection.kind === "ws" ? "正向连接 " : "反向监听 ") + (connection.index + 1) + " · " + connection.url;
    $("connection-select").append(option);
  }
  $("connection-owners").value = (value.superUsers || []).join(", ");
  $("connection-running-owners").textContent = "本次启动主人账号：" + ((value.runningSuperUsers || []).join("、") || "未配置");
  renderSelectedConnection();
}
async function loadConnections() {
  if (connectionBusy || !authenticated) return;
  if (!await confirmConnectionDiscard("重新加载配置")) return;
  const generation = connectionGeneration;
  connectionControls(true); $("connection-error").hidden = true;
  try {
    const value = await api("/api/connections");
    if (generation !== connectionGeneration || !authenticated) return;
    if (typeof resetConnectionAdd === "function") resetConnectionAdd();
    renderConnectionSettings(value);
  }
  catch (error) { if (generation === connectionGeneration && authenticated) { $("connection-save-state").textContent = "配置读取失败，原编辑保留。"; showError("connection-error", error); } }
  finally { if (generation === connectionGeneration) { connectionControls(false); renderConnectionDrafts(); } }
}
$("reload-connections").addEventListener("click", loadConnections);
$("discard-current-connection").addEventListener("click", () => discardConnectionDraft("connection"));
$("discard-connection-owners").addEventListener("click", () => discardConnectionDraft("owners"));
$("connection-select").addEventListener("change", async () => {
  const next = Number($("connection-select").value);
  const generation = connectionGeneration, settings = connectionSettings, index = connectionIndex;
  if (!authenticated || !settings?.connections?.[next] || !await confirmConnectionDiscard("切换连接", true)) {
    if (generation === connectionGeneration && settings === connectionSettings && index === connectionIndex) $("connection-select").value = String(connectionIndex);
    return;
  }
  connectionIndex = next; renderSelectedConnection();
});
$("token-action").addEventListener("change", updateTokenAction);
for (const event of ["input", "change"]) $("connection-form").addEventListener(event, () => { $("connection-discard-status").textContent = ""; renderConnectionDrafts(); });
$("connection-form").addEventListener("submit", async (event) => {
  event.preventDefault(); if (connectionBusy || !connectionSettings?.available || !authenticated) return;
  const generation = connectionGeneration;
  $("connection-error").hidden = true;
  const connection = selectedConnection(), mode = $("token-action").value;
  if (connection && mode === "clear" && !$("clear-token-confirm").checked) { showError("connection-error", new Error("请先确认清除 Token。")); return; }
  const update = { revision:connectionSettings.revision, superUsers:$("connection-owners").value.trim().split(/[\s,，]+/).filter(Boolean) };
  if (connection) {
    update.connection = {kind:connection.kind, index:connection.index, tokenAction:mode};
    if ($("connection-url").value !== connection.url) update.connection.url = $("connection-url").value;
    if (mode === "replace") update.connection.token = $("connection-token").value;
  }
  connectionControls(true);
  try {
    const value = await api("/api/connections", update);
    if (generation !== connectionGeneration || !authenticated) return;
    renderConnectionSettings(value);
    if (value.backupId) $("connection-save-state").textContent += " 旧配置备份已创建。";
  } catch (error) { if (generation === connectionGeneration && authenticated) showError("connection-error", error); }
  finally { if (generation === connectionGeneration) { connectionControls(false); renderConnectionDrafts(); } }
});
document.addEventListener("DOMContentLoaded", async () => {
  try { const state = await api("/api/session"); setupRequired = state.setupRequired; csrf = state.csrfToken || ""; if (state.authenticated) await showWorkspace(); else showAuth(); }
  catch (error) { showAuth(); showError("auth-error", error); $("login-button").disabled = true; }
});

