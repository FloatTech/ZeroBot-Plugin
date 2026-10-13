// Isolated Chrome: synthetic preview metadata only, no live backups or restores.
import assert from 'node:assert/strict';import fs from 'node:fs/promises';import path from 'node:path';import {createRequire} from 'node:module';import test from 'node:test';
const require=createRequire(import.meta.url);const modules=process.env.LOCALADMIN_NODE_MODULES || path.join(process.env.USERPROFILE,'.cache/codex-runtimes/codex-primary-runtime/dependencies/node/node_modules');const {chromium}=require(path.join(modules,'playwright'));
const html=(await fs.readFile('internal/localadmin/assets/index.html','utf8')).replace(/<script\b[^>]*>[\s\S]*?<\/script>/g,'').replace(/<link\b[^>]*>/g,'');const app=await fs.readFile('internal/localadmin/assets/app.js','utf8'),backups=await fs.readFile('internal/localadmin/assets/backups.js','utf8');
async function fixture(t){const browser=await chromium.launch({channel:'chrome',headless:true});t.after(()=>browser.close());const context=await browser.newContext();await context.route('**/*',r=>r.abort());const page=await context.newPage();await page.setContent(html);await page.addScriptTag({content:app});await page.addScriptTag({content:backups});await page.evaluate(()=>{
 authenticated=true;csrf='impact-fixture';window.showOverview=()=>{};window.showLogs=()=>{};document.getElementById('auth').hidden=true;document.getElementById('workspace').hidden=false;
 window.data={backups:['startup-config','control-db','admin-options'].map((scope,i)=>({id:'fixture-'+i,scope,valid:true,integrity:true,version:1,size:512,createdAt:'2026-10-09T00:00:00Z'})),scopes:{'startup-config':true,'control-db':true,'admin-options':true},more:false};
 window.requests=[];window.preview={};window.fetch=async(url,options)=>{requests.push({url,body:options.body?JSON.parse(options.body):null});return {ok:true,status:200,json:async()=>url==='/api/backups/preview'?preview:data};};document.querySelector('.nav-button[data-view=backups]').click();
 });await page.waitForFunction(()=>!backupBusy);return page;}
async function show(page,index,impacts,changed=true,expiresAt=Math.floor(Date.now()/1000)+600){await page.evaluate(({index,impacts,changed,expiresAt})=>{window.preview={backup:data.backups[index],changed,impacts,expiresAt,revision:'snapshot-revision',token:'snapshot-ticket',tables:4,rows:4,changedTables:1,warnings:['fixed scope warning']};}, {index,impacts,changed,expiresAt});await page.locator('#backup-list .backup-card button').nth(index).click();await page.waitForFunction(()=>!backupBusy);}

test('three scopes render count/content differences as literal text; filtering keeps fixed preview and cancel stays readonly',async t=>{
 const page=await fixture(t);const cases=[
 [{label:'正向 OneBot 连接',changed:true,currentCount:2,backupCount:2,unit:'条连接'},{label:'机器人主人',changed:true,currentCount:0,backupCount:1,unit:'个账号'},{label:'其他启动设置',changed:false,currentCount:2,backupCount:2,unit:'项设置'}],
 [{label:'响应规则与AI扩展',changed:true,currentCount:100000,backupCount:100000,unit:'条记录'},{label:'<img src=x onerror=bad()>',changed:false,currentCount:0,backupCount:0,unit:'条记录'}],
 [{label:'后台启用状态',changed:true},{label:'后台访问地址',changed:false}]
 ];
 for(let i=0;i<cases.length;i++){await show(page,i,cases[i]);assert.equal(await page.locator('#restore-impact-list article').count(),cases[i].length);assert.equal(await page.locator('#restore-impact-list img').count(),0);assert.equal(await page.locator('#restore-confirm').isChecked(),false);assert.equal(await page.locator('#restore-backup').isDisabled(),true);assert.match(await page.locator('#restore-impact-status').textContent(),/个类别/);if(i<2)assert.match(await page.locator('#restore-impact-list').textContent(),/数量相同，内容不同/);}
 await page.locator('#restore-confirm').check();assert.equal(await page.locator('#restore-backup').isEnabled(),true);
 const before=await page.locator('#restore-impact-list').textContent();await page.locator('#backup-search').fill('no-match');assert.equal(await page.locator('#restore-impact-list').textContent(),before);assert.equal(await page.evaluate(()=>backupPreview.backup.id),'fixture-2');
 await page.locator('#cancel-restore').click();assert.equal(await page.locator('#restore-impact-list article').count(),0);assert.equal(await page.locator('#restore-impact-status').textContent(),'');assert.equal(await page.locator('#restore-preview').isVisible(),false);
 const requests=await page.evaluate(()=>requests);assert.equal(requests.length,4);assert.deepEqual(requests.filter(r=>r.body).map(r=>r.url),Array(3).fill('/api/backups/preview'));
});

test('equal preview stays nonrestorable; no-metadata fallback works; expiry, refresh and navigation clear old impacts',async t=>{
 const page=await fixture(t);await show(page,0,[{label:'机器人主人',changed:false,currentCount:1,backupCount:1,unit:'个账号'}],false);await page.locator('#restore-confirm').check();assert.equal(await page.locator('#restore-backup').isDisabled(),true);assert.match(await page.locator('#restore-impact-status').textContent(),/0 个有变化/);
 await show(page,2,undefined,true);assert.match(await page.locator('#restore-impact-status').textContent(),/未提供类别摘要/);assert.equal(await page.locator('#restore-impact-list article').count(),0);
 await show(page,0,[{label:'字段名称或表示差异',changed:true}],true,1);assert.equal(await page.locator('#restore-confirm').isDisabled(),true);assert.equal(await page.locator('#restore-confirm').isChecked(),false);assert.equal(await page.locator('#restore-backup').isDisabled(),true);assert.match(await page.locator('#restore-expiry-state').textContent(),/过期/);assert.equal(await page.locator('#restore-impact-list article').count(),1);
 await show(page,0,[{label:'正向 OneBot 连接',changed:true,currentCount:1,backupCount:2,unit:'条连接'}]);await page.locator('#reload-backups').click();await page.waitForFunction(()=>!backupBusy);assert.equal(await page.locator('#restore-impact-list article').count(),0);
 await show(page,0,[{label:'正向 OneBot 连接',changed:true}]);await page.locator('.nav-button[data-view=overview]').click();assert.equal(await page.locator('#restore-impact-list article').count(),0);
 assert.equal((await page.evaluate(()=>requests)).some(r=>r.url==='/api/backups/restore'),false);
});

test('busy preview and failure clear old summary; stale success/401 cannot repopulate new generation or logged-out UI',async t=>{
 const page=await fixture(t);await show(page,0,[{label:'old category',changed:true}]);
 await page.evaluate(()=>{window.pending=[];window.fetch=(url,options)=>{requests.push({url,body:options.body?JSON.parse(options.body):null});return new Promise(resolve=>pending.push({resolve,signal:options.signal}));};});
 await page.locator('#backup-list button').nth(1).click();await page.waitForFunction(()=>pending.length===1);assert.equal(await page.locator('#restore-impact-list article').count(),0);assert.equal(await page.locator('#backup-list button').first().isDisabled(),true);
 await page.evaluate(()=>pending[0].resolve({ok:false,status:409,json:async()=>({error:'changed snapshot'})}));await page.waitForFunction(()=>!backupBusy);assert.equal(await page.locator('#restore-impact-list article').count(),0);assert.equal(await page.locator('#restore-backup').isDisabled(),true);
 await page.locator('#backup-list button').nth(0).click();await page.waitForFunction(()=>pending.length===2);
 await page.evaluate(()=>{showBackups(false);showBackups(true);});await page.waitForFunction(()=>pending.length===3);assert.equal(await page.evaluate(()=>pending[1].signal.aborted),true);
 await page.evaluate(()=>pending[1].resolve({ok:true,status:200,json:async()=>({...preview,impacts:[{label:'STALE',changed:true}]})}));await page.evaluate(()=>pending[2].resolve({ok:true,status:200,json:async()=>data}));await page.waitForFunction(()=>!backupBusy);assert.equal(await page.locator('#restore-impact-list article').count(),0);
 await page.locator('#backup-list button').nth(0).click();await page.waitForFunction(()=>pending.length===4);await page.evaluate(()=>{authenticated=false;showAuth();authenticated=true;csrf='new-impact-fixture';showBackups(true);});await page.waitForFunction(()=>pending.length===5);
 await page.evaluate(()=>pending[3].resolve({ok:false,status:401,json:async()=>({error:'old expired'})}));await page.evaluate(()=>Promise.resolve());assert.equal(await page.evaluate(()=>authenticated),true);assert.equal(await page.evaluate(()=>csrf),'new-impact-fixture');assert.equal(await page.locator('#restore-impact-list article').count(),0);
 await page.evaluate(()=>pending[4].resolve({ok:true,status:200,json:async()=>data}));await page.waitForFunction(()=>!backupBusy);
 await page.evaluate(()=>{authenticated=false;showAuth();});assert.equal(await page.locator('#restore-impact-list article').count(),0);assert.equal(await page.locator('#restore-impact-status').textContent(),'');
});
