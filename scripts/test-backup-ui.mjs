// Offline behavior tests; never attaches to a browser or reads local credentials.
import assert from 'node:assert/strict';
import fs from 'node:fs';
import vm from 'node:vm';
import test from 'node:test';

class Element {
  constructor(tag='div'){this.tagName=tag;this.children=[];this.listeners=new Map();this.dataset={};this.value='';this.hidden=false;this.disabled=false;this.checked=false;this._text='';}
  set textContent(value){this._text=String(value);this.replaceChildren();} get textContent(){return this._text;}
  set innerHTML(value){throw new Error('backup content must remain text');}
  append(...items){this.children.push(...items);} replaceChildren(...items){this.children=[];this.append(...items);}
  get firstElementChild(){return this.children[0];}
  querySelectorAll(tag){return this.children.flatMap(item=>[...(item.tagName===tag?[item]:[]),...item.querySelectorAll(tag)]);}
  addEventListener(event,callback){this.listeners.set(event,callback);} dispatch(event){return this.listeners.get(event)?.({preventDefault(){}});}
}
function harness(){
  const elements=new Map(),requests=[];
  const get=id=>{if(!elements.has(id))elements.set(id,new Element());return elements.get(id);};
  const context=vm.createContext({document:{createElement:tag=>new Element(tag),addEventListener(){}},setTimeout:()=>1,clearTimeout(){},$:get,authenticated:true,AbortController,Date,
    connectionSettings:{},plugins:[{}],pluginDrafts:new Map(),pluginListRequest:0,pluginGeneration:0,pluginSaving:false,pluginScopeBusy:false,pluginBatch:null,pluginControlAvailable:false,pluginResponse:null,pluginResponseLoaded:false,pluginReloadAfterRestore:false,renderPluginBusy(){},renderPlugins(){},resetAI(){},catalogResets:0,resetPluginCatalog(){vm.runInContext('catalogResets++',context);},
    showError:(id,error)=>{get(id).textContent=error.message;get(id).hidden=false;},
    api:(path,body,signal)=>new Promise((resolve,reject)=>{requests.push({path,body,signal,resolve,reject});signal?.addEventListener('abort',()=>reject(Object.assign(new Error('aborted'),{name:'AbortError'})),{once:true});})});
  vm.runInContext(fs.readFileSync(new URL('../internal/localadmin/assets/backups.js',import.meta.url),'utf8'),context);
  return {context,get,requests,run:code=>vm.runInContext(code,context)};
}
const flush=()=>new Promise(resolve=>setImmediate(resolve));
const entry={id:'fixture-safe-id',scope:'control-db',valid:true,integrity:true,version:1,size:512,createdAt:'2026-10-02T10:00:00Z'};
const list={backups:[entry],scopes:{'control-db':true},more:false};
const preview={backup:entry,changed:true,revision:'revision',token:'ticket',expiresAt:Math.floor(Date.now()/1000)+600,tables:3,rows:6,changedTables:1,warnings:['<script>literal warning</script>']};

test('backup list and preview use text; restoration requires explicit confirmation',async()=>{
  const h=harness();h.run('showBackups(true)');h.requests[0].resolve(list);await flush();
  assert.equal(h.get('backup-list').children.length,1);
  h.run('previewBackup("fixture-safe-id")');h.requests[1].resolve(preview);await flush();
  assert.equal(h.get('restore-warnings').firstElementChild.textContent,preview.warnings[0]);
  assert.equal(h.get('restore-backup').disabled,true);
  await h.get('restore-backup').dispatch('click');assert.equal(h.requests.length,2);
  h.get('restore-confirm').checked=true;h.get('restore-confirm').dispatch('change');assert.equal(h.get('restore-backup').disabled,false);
  const promise=h.get('restore-backup').dispatch('click');
  assert.equal(h.requests[2].path,'/api/backups/restore');assert.equal(h.requests[2].body.confirm,true);assert.equal(h.requests[2].body.token,'ticket');
  h.requests[2].resolve({changed:true,restartRequired:true,backupId:'safety-backup'});await flush();
  assert.equal(h.run('connectionSettings'),null);assert.equal(h.run('pluginDrafts.size'),0);assert.equal(h.requests[3].path,'/api/backups');
  assert.equal(h.run('catalogResets'),1);
  h.requests[3].resolve(list);await promise;assert.match(h.get('backup-status').textContent,/恢复完成.*重启/);
});

test('navigation and logout cancel reads and refuse stale previews',async()=>{
  const h=harness();h.run('showBackups(true)');h.run('showBackups(false)');assert.equal(h.requests[0].signal.aborted,true);await flush();
  h.run('showBackups(true)');h.requests[1].resolve(list);await flush();
  h.run('previewBackup("fixture-safe-id")');h.run('authenticated=false;resetBackups()');assert.equal(h.requests[2].signal.aborted,true);h.requests[2].resolve(preview);await flush();
  assert.equal(h.get('restore-preview').hidden,true);assert.equal(h.get('backup-list').children.length,0);assert.equal(h.run('backupPreview'),null);
});

test('expired previews and failures cannot be resubmitted without preview',async()=>{
  const h=harness();h.run('showBackups(true)');h.requests[0].resolve(list);await flush();
  h.run('previewBackup("fixture-safe-id")');h.requests[1].resolve({...preview,expiresAt:1});await flush();
  h.get('restore-confirm').checked=true;await h.get('restore-backup').dispatch('click');assert.equal(h.requests.length,2);assert.match(h.get('backup-error').textContent,/过期/);
  h.run('previewBackup("fixture-safe-id")');h.requests[2].resolve(preview);await flush();h.get('restore-confirm').checked=true;
  const pending=h.get('restore-backup').dispatch('click');h.requests[3].reject(new Error('配置已变化'));await pending;
  assert.equal(h.run('backupPreview'),null);assert.equal(h.get('restore-backup').disabled,true);assert.match(h.get('backup-error').textContent,/配置已变化/);
  assert.equal(h.requests.length,4);
});

const cleanupPreview={policy:{keepPerScope:10,keepDays:7},candidates:[entry],bytes:512,skipped:2,more:false,revision:'cleanup-revision',asOf:Math.floor(Date.now()/1000),expiresAt:Math.floor(Date.now()/1000)+600,token:'cleanup-ticket'};
async function cleanupHarness(){
  const h=harness();h.get('cleanup-keep').value='10';h.get('cleanup-days').value='7';h.run('showBackups(true)');h.requests[0].resolve(list);await flush();return h;
}

test('cleanup previews remain text and deletion requires explicit confirmation',async()=>{
  const h=await cleanupHarness();const pending=h.get('cleanup-form').dispatch('submit');
  assert.equal(h.requests[1].path,'/api/backups/cleanup-preview');assert.equal(h.requests[1].body.keepPerScope,10);
  h.requests[1].resolve({...cleanupPreview,candidates:[{...entry,id:'<script>literal</script>'}]});await pending;
  assert.match(h.get('cleanup-list').firstElementChild.textContent,/<script>literal<\/script>/);
  assert.match(h.get('cleanup-detail').textContent,/跳过 2/);assert.equal(h.get('cleanup-backups').disabled,true);
  await h.get('cleanup-backups').dispatch('click');assert.equal(h.requests.length,2);
  h.get('cleanup-confirm').checked=true;h.get('cleanup-confirm').dispatch('change');assert.equal(h.get('cleanup-backups').disabled,false);
  const deletion=h.get('cleanup-backups').dispatch('click');assert.equal(h.requests[2].path,'/api/backups/cleanup');assert.equal(h.requests[2].body.confirm,true);assert.equal(h.requests[2].body.token,'cleanup-ticket');
  h.requests[2].resolve({deleted:[entry.id],partial:false,message:'清理完成'});await flush();assert.equal(h.run('backupCleanupPreview'),null);
  h.requests[3].resolve(list);await deletion;assert.match(h.get('backup-status').textContent,/完整删除 1/);assert.equal(h.get('cleanup-backups').disabled,true);
});

test('cleanup policy edits, navigation and logout discard confirmation tickets',async()=>{
  const h=await cleanupHarness();h.get('cleanup-form').dispatch('submit');h.requests[1].resolve(cleanupPreview);await flush();
  h.get('cleanup-confirm').checked=true;h.get('cleanup-days').value='8';h.get('cleanup-days').dispatch('input');
  assert.equal(h.run('backupCleanupPreview'),null);assert.equal(h.get('cleanup-confirm').checked,false);assert.equal(h.get('cleanup-backups').disabled,true);
  h.get('cleanup-form').dispatch('submit');h.run('showBackups(false)');assert.equal(h.requests[2].signal.aborted,true);h.requests[2].resolve(cleanupPreview);await flush();assert.equal(h.get('cleanup-preview').hidden,true);
  h.run('showBackups(true)');h.requests[3].resolve(list);await flush();h.get('cleanup-form').dispatch('submit');
  h.run('authenticated=false;resetBackups()');assert.equal(h.requests[4].signal.aborted,true);h.requests[4].resolve(cleanupPreview);await flush();assert.equal(h.run('backupCleanupPreview'),null);
});

test('expired/conflicting cleanup and partial failure cannot silently retry',async()=>{
  const h=await cleanupHarness();h.get('cleanup-form').dispatch('submit');h.requests[1].resolve({...cleanupPreview,expiresAt:1});await flush();h.get('cleanup-confirm').checked=true;
  await h.get('cleanup-backups').dispatch('click');assert.equal(h.requests.length,2);assert.match(h.get('backup-error').textContent,/过期/);
  h.get('cleanup-form').dispatch('submit');h.requests[2].resolve(cleanupPreview);await flush();h.get('cleanup-confirm').checked=true;
  const failed=h.get('cleanup-backups').dispatch('click');h.requests[3].reject(new Error('备份清单已变化'));await failed;assert.equal(h.run('backupCleanupPreview'),null);assert.match(h.get('backup-error').textContent,/已变化/);
  h.get('cleanup-form').dispatch('submit');h.requests[4].resolve(cleanupPreview);await flush();h.get('cleanup-confirm').checked=true;
  const partial=h.get('cleanup-backups').dispatch('click');h.requests[5].resolve({deleted:[],partial:true,message:'部分文件可能已删除，请刷新'});await flush();h.requests[6].resolve(list);await partial;
  assert.match(h.get('backup-error').textContent,/部分文件/);assert.equal(h.run('backupCleanupPreview'),null);assert.equal(h.get('cleanup-confirm').checked,false);
});

test('empty and invalid cleanup plans never submit a deletion',async()=>{
  const h=await cleanupHarness();h.get('cleanup-keep').value='2';await h.get('cleanup-form').dispatch('submit');assert.equal(h.requests.length,1);assert.match(h.get('backup-error').textContent,/3～200/);
  h.get('cleanup-keep').value='10';h.get('cleanup-form').dispatch('submit');h.requests[1].resolve({...cleanupPreview,candidates:[]});await flush();
  h.get('cleanup-confirm').checked=true;h.get('cleanup-confirm').dispatch('change');assert.equal(h.get('cleanup-backups').disabled,true);await h.get('cleanup-backups').dispatch('click');assert.equal(h.requests.length,2);
});

const searchEntries=[
  {...entry,id:'20261009T000000Z-control'},
  {...entry,id:'20261008T000000Z-startup',scope:'startup-config',integrity:false,createdAt:'2026-10-08T00:00:00Z'},
  {...entry,id:'20261007T000000Z-admin',scope:'admin-options',createdAt:'2026-10-07T00:00:00Z'},
  {id:'<img src=x onerror=bad()>',valid:false}
];
const searchList={backups:searchEntries,scopes:{'control-db':true,'startup-config':true,'admin-options':true},more:true};
async function searchHarness(){const h=harness();h.run('showBackups(true)');h.requests[0].resolve(searchList);await flush();return h;}
const shownIDs=h=>h.get('backup-list').children.filter(c=>c.className==='panel backup-card').map(c=>c.children[0].children[2].textContent);

test('backup search combines metadata/scope/status with loaded counts and keeps unsafe text literal',async()=>{
  const h=await searchHarness();assert.deepEqual(shownIDs(h),searchEntries.map(e=>e.id));
  assert.match(h.get('backup-filter-count').textContent,/4 \/ 4.*含摘要 2.*旧版无摘要 1.*不可恢复 1/);
  assert.match(h.get('backup-filter-count').textContent,/更早未加载/);
  h.get('backup-scope-filter').value='startup-config';h.get('backup-scope-filter').dispatch('change');
  h.get('backup-state-filter').value='legacy';h.get('backup-state-filter').dispatch('change');
  assert.deepEqual(shownIDs(h),[searchEntries[1].id]);
  h.get('backup-search').value='2026-10-08';h.get('backup-search').dispatch('input');assert.deepEqual(shownIDs(h),[searchEntries[1].id]);
  h.get('backup-state-filter').value='verified';h.get('backup-state-filter').dispatch('change');
  assert.equal(shownIDs(h).length,0);assert.match(h.get('backup-list').firstElementChild.textContent,/没有符合/);
  h.get('clear-backup-filter').dispatch('click');h.get('backup-search').value='后台启用';h.get('backup-search').dispatch('input');assert.deepEqual(shownIDs(h),[searchEntries[2].id]);
  h.get('clear-backup-filter').dispatch('click');h.get('backup-scope-filter').value='unknown';h.get('backup-scope-filter').dispatch('change');
  assert.deepEqual(shownIDs(h),[searchEntries[3].id]);assert.equal(h.get('backup-list').firstElementChild.children[1].disabled,true);
  assert.equal(h.requests.length,1); // Filtering never calls any read or write API.
  h.context.emptyList={backups:[],scopes:{},more:false};h.run('renderBackups(emptyList)');assert.match(h.get('backup-list').firstElementChild.textContent,/暂无备份/);
});

test('hiding a selected backup does not alter its visible preview identity or restoration request',async()=>{
  const h=await searchHarness();h.run('previewBackup("fixture-safe-id")');h.requests[1].resolve(preview);await flush();
  assert.equal(h.get('restore-backup-id').textContent,'备份编号：fixture-safe-id');h.get('restore-confirm').checked=true;h.get('restore-confirm').dispatch('change');
  h.get('backup-search').value='no-match';h.get('backup-search').dispatch('input');
  assert.equal(shownIDs(h).length,0);assert.equal(h.get('restore-preview').hidden,false);assert.equal(h.get('restore-confirm').checked,true);
  assert.equal(h.run('backupPreview.backup.id'),'fixture-safe-id');assert.equal(h.requests.length,2);
  const restoring=h.get('restore-backup').dispatch('click');assert.equal(h.requests[2].body.id,'fixture-safe-id');assert.equal(h.requests[2].body.token,'ticket');
  h.requests[2].reject(new Error('模拟恢复拒绝'));await restoring;assert.equal(h.get('restore-backup-id').textContent,'');
});

test('filters survive refresh/navigation/create, failed reads mark stale counts and session reset clears cache',async()=>{
  const h=await searchHarness();h.get('backup-state-filter').value='legacy';h.get('backup-state-filter').dispatch('change');
  h.get('reload-backups').dispatch('click');assert.equal(h.get('backup-state-filter').disabled,true);
  h.get('clear-backup-filter').dispatch('click');assert.equal(h.run('backupStateFilter'),'legacy');
  h.requests[1].reject(new Error('模拟读取失败'));await flush();assert.deepEqual(shownIDs(h),[searchEntries[1].id]);assert.match(h.get('backup-filter-count').textContent,/上次读取结果.*刷新失败/);
  h.run('showBackups(false);showBackups(true)');h.requests[2].resolve(searchList);await flush();assert.equal(h.run('backupStateFilter'),'legacy');assert.doesNotMatch(h.get('backup-filter-count').textContent,/刷新失败/);
  const creating=h.get('backup-create-form').dispatch('submit');h.requests[3].resolve({});await flush();
  const fresh={...searchList,backups:[...searchEntries,{...searchEntries[1],id:'new-legacy'}]};h.requests[4].resolve(fresh);await creating;
  assert.equal(shownIDs(h).length,2);assert.equal(h.run('backupStateFilter'),'legacy');
  h.run('authenticated=false;resetBackups()');assert.equal(h.run('backupCatalog'),null);assert.equal(h.run('backupSearch'),'');assert.equal(h.get('backup-state-filter').value,'all');assert.equal(h.get('backup-list').children.length,0);
});

test('old resolved catalogue cannot repopulate a reset session and filters cannot send requests while unauthenticated',async()=>{
  const h=harness();h.run('showBackups(true)');h.requests[0].resolve(searchList);
  h.run('authenticated=false;resetBackups();authenticated=true;showBackups(true)');await flush();
  assert.equal(h.run('backupCatalog'),null);assert.equal(h.get('backup-list').children.length,0);assert.equal(h.get('backup-state-filter').disabled,true);
  h.requests[1].resolve(searchList);await flush();assert.equal(shownIDs(h).length,4);
  h.run('authenticated=false;resetBackups()');h.get('backup-search').value='private draft';h.get('backup-search').dispatch('input');
  h.get('clear-backup-filter').dispatch('click');assert.equal(h.requests.length,2);assert.equal(h.run('backupSearch'),'');assert.equal(h.run('backupCatalog'),null);
});

const pagedList=count=>({backups:Array.from({length:count},(_,i)=>({...entry,id:'ordered-'+String(i).padStart(3,'0'),integrity:i%2===0})),scopes:{'control-db':true},more:count===200});
async function pagingHarness(count){const h=harness();h.run('showBackups(true)');h.requests[0].resolve(pagedList(count));await flush();return h;}

test('backup pagination preserves full loaded order without duplicates and filters before slicing',async()=>{
  const h=await pagingHarness(200),seen=[];
  assert.equal(shownIDs(h).length,20);assert.equal(h.run('backupPageCount'),10);assert.equal(h.get('backup-prev').disabled,true);
  h.get('backup-prev').dispatch('click');assert.equal(h.run('backupPage'),1);
  for(let i=1;i<=10;i++){assert.equal(h.run('backupPage'),i);seen.push(...shownIDs(h));h.get('backup-next').dispatch('click');}
  assert.deepEqual(seen,pagedList(200).backups.map(e=>e.id));assert.equal(new Set(seen).size,200);assert.equal(h.get('backup-next').disabled,true);
  assert.match(h.get('backup-page-status').textContent,/第 10 \/ 10 页.*181～200 \/ 200/);
  h.get('backup-page-size').value='10';h.get('backup-page-size').dispatch('change');assert.equal(h.run('backupPage'),1);assert.equal(h.run('backupPageCount'),20);assert.equal(shownIDs(h).length,10);
  h.get('backup-page-size').value='50';h.get('backup-page-size').dispatch('change');assert.equal(h.run('backupPageCount'),4);assert.equal(shownIDs(h).length,50);
  h.get('backup-next').dispatch('click');h.get('backup-state-filter').value='legacy';h.get('backup-state-filter').dispatch('change');
  assert.equal(h.run('backupPage'),1);assert.equal(h.run('backupMatchCount'),100);assert.equal(shownIDs(h).length,50);assert.ok(shownIDs(h).every(id=>Number(id.slice(-3))%2===1));
  h.get('backup-search').value='absent';h.get('backup-search').dispatch('input');assert.equal(h.get('backup-prev').disabled,true);assert.equal(h.get('backup-next').disabled,true);assert.match(h.get('backup-page-status').textContent,/没有匹配/);
  h.get('clear-backup-filter').dispatch('click');assert.equal(h.run('backupPage'),1);assert.equal(h.run('backupPageSize'),50);assert.equal(h.requests.length,1);
});

test('page state survives failed/successful refresh and navigation, clamps shrink, freezes while busy and resets on logout',async()=>{
  const h=await pagingHarness(65);for(let i=0;i<3;i++)h.get('backup-next').dispatch('click');assert.equal(h.run('backupPage'),4);assert.equal(shownIDs(h).length,5);
  h.get('reload-backups').dispatch('click');assert.equal(h.get('backup-prev').disabled,true);assert.equal(h.get('backup-page-size').disabled,true);
  h.get('backup-prev').dispatch('click');assert.equal(h.run('backupPage'),4);
  h.requests[1].reject(new Error('模拟读取失败'));await flush();assert.equal(h.run('backupPage'),4);assert.match(h.get('backup-filter-count').textContent,/刷新失败/);
  h.get('reload-backups').dispatch('click');h.requests[2].resolve(pagedList(25));await flush();assert.equal(h.run('backupPage'),2);assert.equal(shownIDs(h).length,5);
  assert.match(h.get('backup-page-status').textContent,/21～25 \/ 25/);
  h.run('showBackups(false);showBackups(true)');h.requests[3].resolve(pagedList(25));await flush();assert.equal(h.run('backupPage'),2);
  const creating=h.get('backup-create-form').dispatch('submit');h.requests[4].resolve({});await flush();h.requests[5].resolve(pagedList(30));await creating;assert.equal(h.run('backupPage'),2);assert.equal(shownIDs(h).length,10);
  h.get('backup-page-size').value='10';h.get('backup-page-size').dispatch('change');h.get('backup-next').dispatch('click');
  h.run('authenticated=false;resetBackups()');assert.equal(h.run('backupPage'),1);assert.equal(h.run('backupPageSize'),20);assert.equal(h.get('backup-page-size').value,'20');assert.equal(h.get('backup-next').disabled,true);assert.equal(h.run('backupCatalog'),null);
  const count=h.requests.length;h.get('backup-next').dispatch('click');assert.equal(h.requests.length,count);
});
