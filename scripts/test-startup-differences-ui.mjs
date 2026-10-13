// Offline Chrome, synthetic settings and mocked fetch only.
import assert from 'node:assert/strict';
import fs from 'node:fs/promises';
import path from 'node:path';
import {createRequire} from 'node:module';
import test,{after} from 'node:test';
const require=createRequire(import.meta.url);
const modules=process.env.LOCALADMIN_NODE_MODULES||path.join(process.env.USERPROFILE,'.cache/codex-runtimes/codex-primary-runtime/dependencies/node/node_modules');
const {chromium}=require(path.join(modules,'playwright'));
const browser=await chromium.launch({channel:'chrome',headless:true});after(()=>browser.close());
const html=(await fs.readFile('internal/localadmin/assets/index.html','utf8')).replace(/<script\b[^>]*>[\s\S]*?<\/script>/g,'').replace(/<link\b[^>]*>/g,'');
const scripts=await Promise.all(['app.js','connections-add.js','connections-delete.js'].map(f=>fs.readFile('internal/localadmin/assets/'+f,'utf8')));
const data=()=>{const connections=[{kind:'ws',index:0,url:'ws://localhost:17000',tokenConfigured:true}];return {available:true,revision:'original',connections,runningConnections:connections,superUsers:['12345678'],runningSuperUsers:['12345678'],restartRequired:false,startupDifferences:['正向 OneBot 连接','反向 OneBot 连接','机器人主人','其他机器人参数','其他启动设置'].map((label,i)=>({label,changed:false,runningCount:i===0||i===2?1:0,savedCount:i===0||i===2?1:0,unit:i<2?'条连接':i===2?'个账号':'项设置'}))};};
const changed=()=>{const d=data();d.revision='secret-changed';d.restartRequired=true;d.startupDifferences[0].changed=true;return d;};
async function fixture(t){const context=await browser.newContext();t.after(()=>context.close());await context.route('**/*',r=>r.abort());const p=await context.newPage();await p.setContent(html);for(const content of scripts)await p.addScriptTag({content});await p.evaluate(d=>{
 authenticated=true;csrf='startup-session';window.showOverview=()=>{};window.showLogs=()=>{};window.showBackups=()=>{};document.getElementById('auth').hidden=true;document.getElementById('workspace').hidden=false;for(const v of document.querySelectorAll('.content>section'))v.hidden=v.id!=='connections-view';window.requests=[];window.replies=[];window.pending=[];window.fetch=async(url,options)=>{requests.push({url,body:options.body?JSON.parse(options.body):null});const v=replies.shift();if(!v)return new Promise(resolve=>pending.push(resolve));return {ok:!v.error,status:v.status||200,json:async()=>v};};renderConnectionSettings(d);connectionControls(false);
 },data());return p;}
const idle=p=>p.waitForFunction(()=>!connectionBusy);
const text=p=>p.locator('#startup-difference-list').textContent();

test('summary renders same-count changes and optional representation row as literal text, hides unavailable and supports older API',async t=>{
 const p=await fixture(t);assert.equal(await p.locator('.startup-difference-row').count(),5);assert.match(await text(p),/本次启动 0条连接 → 已保存 0条连接/);assert.match(await p.locator('#startup-difference-status').textContent(),/一致.*无需/);
 const d=changed();d.startupDifferences.push({label:'字段名称或表示差异',changed:true});await p.evaluate(d=>renderConnectionSettings(d),d);assert.match(await text(p),/数量相同，内容不同/);assert.match(await p.locator('#startup-difference-status').textContent(),/待重启；2 个类别/);assert.equal(await p.locator('.startup-difference-row').last().locator('p').count(),0);
 d.startupDifferences[0].label='<img src=x onerror=bad()>';d.startupDifferences[0].unit='<script>bad()</script>';await p.evaluate(d=>renderConnectionSettings(d),d);assert.equal(await p.locator('#startup-difference-list img,#startup-difference-list script').count(),0);assert.match(await text(p),/<img/);
 await p.evaluate(d=>{delete d.startupDifferences;renderConnectionSettings(d);},data());assert.match(await p.locator('#startup-difference-status').textContent(),/暂无分类摘要/);await p.evaluate(()=>renderConnectionSettings({available:false,connections:[],superUsers:[]}));assert.equal(await p.locator('#startup-differences').isHidden(),true);assert.equal(await text(p),'');assert.equal(await p.evaluate(()=>requests.length),0);
});

test('draft edits and navigation do not change saved summary; failed reload retains last result and successful reload updates it',async t=>{
 const p=await fixture(t);const original=await text(p);await p.locator('#connection-url').fill('ws://localhost:19000');await p.selectOption('#token-action','replace');await p.locator('#connection-token').fill('private-draft-token');await p.locator('#connection-owners').fill('87654321');assert.equal(await text(p),original);assert.doesNotMatch(await text(p),/19000|private-draft-token|87654321/);await p.locator('[data-view=overview]').click();await p.locator('[data-view=connections]').click();assert.equal(await text(p),original);assert.equal(await p.evaluate(()=>requests.length),0);
 await p.evaluate(()=>replies.push({error:'读取失败',status:500}));await p.locator('#reload-connections').click();await p.locator('#connection-draft-confirm-accept').click();await idle(p);assert.equal(await text(p),original);assert.equal(await p.locator('#connection-token').inputValue(),'private-draft-token');assert.match(await p.locator('#connection-save-state').textContent(),/读取失败/);
 await p.evaluate(d=>replies.push(d),changed());await p.locator('#reload-connections').click();await p.locator('#connection-draft-confirm-accept').click();await idle(p);assert.match(await text(p),/数量相同，内容不同/);assert.equal(await p.locator('#connection-token').inputValue(),'');assert.match(await p.locator('#startup-difference-status').textContent(),/待重启/);
});

test('failed save retains summary and draft; successful edit, addition and deletion refresh it from server snapshots',async t=>{
 const p=await fixture(t);await p.locator('#connection-owners').fill('12345678, 23456789');const original=await text(p);await p.evaluate(()=>replies.push({error:'保存失败',status:500}));await p.locator('#save-connections').click();await idle(p);assert.equal(await text(p),original);assert.equal(await p.locator('#connection-owners').inputValue(),'12345678, 23456789');
 const owner=data();owner.superUsers.push('23456789');owner.revision='owners';owner.backupId='owner-backup';owner.restartRequired=true;owner.startupDifferences[2].changed=true;owner.startupDifferences[2].savedCount=2;await p.evaluate(d=>replies.push(d),owner);await p.locator('#save-connections').click();await idle(p);assert.match(await text(p),/本次启动 1个账号 → 已保存 2个账号/);
 const added=structuredClone(owner);added.revision='added';added.backupId='add-backup';added.connections.push({kind:'ws',index:1,url:'ws://localhost:17777',tokenConfigured:true});added.startupDifferences[0].changed=true;added.startupDifferences[0].savedCount=2;await p.locator('#connection-add-url').fill('ws://localhost:17777');await p.locator('#connection-add-token').fill('private-add-token');await p.evaluate(d=>replies.push(d),added);await p.locator('#connection-add-save').click();await p.locator('#connection-add-confirm-accept').click();await idle(p);assert.match(await text(p),/本次启动 1条连接 → 已保存 2条连接/);assert.match(await p.locator('#startup-difference-status').textContent(),/2 个类别/);
 const removed=structuredClone(owner);removed.revision='removed';removed.backupId='remove-backup';await p.evaluate(d=>replies.push(d),removed);await p.locator('#preview-connection-delete').click();await p.locator('#connection-delete-confirm').check();await p.locator('#connection-delete-save').click();await idle(p);assert.match(await p.locator('#startup-difference-status').textContent(),/1 个类别/);assert.doesNotMatch(await text(p),/private-add-token/);
});

test('expiry clears summary and late reads cannot overwrite a new authenticated session',async t=>{
 const p=await fixture(t);await p.locator('#reload-connections').click();await p.waitForFunction(()=>pending.length===1);await p.evaluate(()=>{authenticated=false;csrf='';showAuth();});assert.equal(await p.locator('#startup-differences').isHidden(),true);assert.equal(await text(p),'');
 await p.evaluate(d=>{authenticated=true;csrf='new-session';document.getElementById('auth').hidden=true;document.getElementById('workspace').hidden=false;renderConnectionSettings(d);connectionControls(false);},changed());await p.locator('#reload-connections').click();await p.waitForFunction(()=>pending.length===2);await p.evaluate(d=>pending[0]({ok:true,status:200,json:async()=>d}),data());await p.waitForTimeout(80);assert.match(await p.locator('#startup-difference-status').textContent(),/待重启/);assert.equal(await p.evaluate(()=>connectionBusy),true);
 await p.evaluate(()=>pending[1]({ok:false,status:401,json:async()=>({error:'expired'})}));await p.waitForFunction(()=>!authenticated);assert.equal(await text(p),'');assert.equal(await p.locator('#startup-differences').isHidden(),true);
});
