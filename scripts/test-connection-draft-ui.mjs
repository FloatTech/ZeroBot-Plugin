// Offline Chrome: synthetic connection settings and mocked APIs, no real tokens or QQ.
import assert from 'node:assert/strict';
import fs from 'node:fs/promises';
import path from 'node:path';
import {createRequire} from 'node:module';
import test, {after} from 'node:test';
const require=createRequire(import.meta.url);
const modules=process.env.LOCALADMIN_NODE_MODULES || path.join(process.env.USERPROFILE,'.cache/codex-runtimes/codex-primary-runtime/dependencies/node/node_modules');
const {chromium}=require(path.join(modules,'playwright'));
const html=(await fs.readFile('internal/localadmin/assets/index.html','utf8')).replace(/<script\b[^>]*>[\s\S]*?<\/script>/g,'').replace(/<link\b[^>]*>/g,'');
const app=await fs.readFile('internal/localadmin/assets/app.js','utf8');
const browser=await chromium.launch({channel:'chrome',headless:true});
after(()=>browser.close());
function fixture(){return {available:true,revision:'a'.repeat(64),connections:[{kind:'ws',index:0,url:'ws://127.0.0.1:16700',tokenConfigured:true},{kind:'wss',index:0,url:'ws://127.0.0.1:16701',tokenConfigured:false}],superUsers:['12345678'],runningSuperUsers:['12345678'],runningConnections:[{kind:'ws',index:0,url:'ws://127.0.0.1:16700',tokenConfigured:true}],restartRequired:false};}
async function harness(t){
  const context=await browser.newContext();t.after(()=>context.close());await context.route('**/*',r=>r.abort());
  const page=await context.newPage();await page.setContent(html);await page.addScriptTag({content:app});
  await page.evaluate(data=>{
    authenticated=true;csrf='fixture-session';window.showOverview=()=>{};window.showLogs=()=>{};window.showBackups=()=>{};
    document.getElementById('auth').hidden=true;document.getElementById('workspace').hidden=false;
    for(const view of document.querySelectorAll('.content>section'))view.hidden=view.id!=='connections-view';
    window.requests=[];window.replies=[];window.pending=[];
    window.fetch=async(url,options)=>{
      requests.push({url,body:options.body?JSON.parse(options.body):null});const reply=replies.shift();
      if(!reply)return new Promise(resolve=>pending.push(resolve));
      return {ok:!reply.error,status:reply.status||200,json:async()=>reply};
    };
    renderConnectionSettings(data);connectionControls(false);
  },fixture());return page;
}
const dirty=page=>page.evaluate(()=>connectionDirty());
const idle=page=>page.waitForFunction(()=>!connectionBusy);
async function confirmAction(page,accept,action){await action();await page.locator('#connection-draft-confirmation').waitFor({state:'visible'});const message=await page.locator('#connection-draft-confirm-message').textContent();assert.match(message,/未保存/);await page.locator(accept?'#connection-draft-confirm-accept':'#connection-draft-confirm-cancel').click();await page.locator('#connection-draft-confirmation').waitFor({state:'hidden'});return message;}
const showSession=(page,token)=>page.evaluate(({data,token})=>{authenticated=false;csrf='';showAuth();authenticated=true;csrf=token;document.getElementById('auth').hidden=true;document.getElementById('workspace').hidden=false;renderConnectionSettings(data);connectionControls(false);},{data:fixture(),token});

test('tracks reverted address/owners/token modes, keeps drafts across pages and never includes secrets in status',async t=>{
  const page=await harness(t);assert.equal(await dirty(page),false);
  await page.locator('#connection-url').fill('ws://127.0.0.1:16702');assert.equal(await dirty(page),true);
  await page.locator('#connection-url').fill(fixture().connections[0].url);assert.equal(await dirty(page),false);
  await page.locator('#connection-owners').fill('23456789');assert.match(await page.locator('#connection-draft-status').textContent(),/主人账号/);
  await page.locator('#connection-owners').fill('12345678');assert.equal(await dirty(page),false);
  await page.selectOption('#token-action','replace');await page.locator('#connection-token').fill('synthetic-secret-token');
  assert.equal(await page.locator('#connection-token').getAttribute('type'),'password');assert.doesNotMatch(await page.locator('#connection-draft-status').textContent(),/synthetic-secret-token/);
  await page.locator('[data-view="plugins"]').click();await page.locator('[data-view="connections"]').click();
  assert.equal(await page.locator('#connection-token').inputValue(),'synthetic-secret-token');assert.equal(await page.evaluate(()=>requests.length),0);
  await page.selectOption('#token-action','keep');assert.equal(await page.locator('#connection-token').inputValue(),'');assert.equal(await dirty(page),false);
});

test('cancelled reload/logout/connection switch do not send requests; failed reads preserve values and revision',async t=>{
  const page=await harness(t);await page.locator('#connection-url').fill('ws://127.0.0.1:16702');
  await confirmAction(page,false,()=>page.locator('#reload-connections').click());
  await confirmAction(page,false,()=>page.locator('#logout').click());
  await confirmAction(page,false,()=>page.selectOption('#connection-select','1'));
  assert.equal(await page.locator('#connection-select').inputValue(),'0');assert.equal(await page.evaluate(()=>requests.length),0);
  await page.evaluate(()=>replies.push({error:'模拟读取故障',status:500}));
  await confirmAction(page,true,()=>page.locator('#reload-connections').click());await idle(page);
  assert.equal(await page.locator('#connection-url').inputValue(),'ws://127.0.0.1:16702');assert.equal(await dirty(page),true);
  assert.equal(await page.evaluate(()=>connectionSettings.revision),'a'.repeat(64));assert.equal(await page.locator('#connection-url').isDisabled(),false);
  await page.evaluate(data=>replies.push(data),fixture());await confirmAction(page,true,()=>page.locator('#reload-connections').click());await idle(page);
  assert.equal(await dirty(page),false);assert.equal(await page.locator('#connection-url').inputValue(),fixture().connections[0].url);
});

test('switches only connection drafts while keeping owners, and saves the loaded connection rather than an uncommitted selector',async t=>{
  const page=await harness(t);await page.locator('#connection-owners').fill('23456789');
  await page.selectOption('#connection-select','1');assert.equal(await page.locator('#connection-owners').inputValue(),'23456789');assert.equal(await dirty(page),true);
  await page.locator('#connection-url').fill('ws://127.0.0.1:16703');
  await confirmAction(page,true,()=>page.selectOption('#connection-select','0'));
  assert.equal(await page.locator('#connection-url').inputValue(),fixture().connections[0].url);assert.equal(await page.locator('#connection-owners').inputValue(),'23456789');
  await page.selectOption('#connection-select','1');await page.locator('#connection-url').fill('ws://127.0.0.1:16704');
  await page.evaluate(()=>document.getElementById('connection-select').value='0');
  const saved=fixture();saved.connections[1].url='ws://127.0.0.1:16704';saved.superUsers=['23456789'];saved.restartRequired=true;
  await page.evaluate(data=>replies.push(data),saved);await page.locator('#save-connections').click();await idle(page);
  const [request]=await page.evaluate(()=>requests);assert.deepEqual(request.body,{revision:'a'.repeat(64),superUsers:['23456789'],connection:{kind:'wss',index:0,tokenAction:'keep',url:'ws://127.0.0.1:16704'}});
  assert.equal(await page.locator('#connection-select').inputValue(),'1');assert.equal(await dirty(page),false);assert.match(await page.locator('#connection-save-state').textContent(),/待重启/);
});

test('conflicts keep token and drafts, clear requires confirmation and a successful save clears replacement values',async t=>{
  const page=await harness(t);await page.selectOption('#token-action','clear');await page.locator('#save-connections').click();
  assert.equal(await page.evaluate(()=>requests.length),0);assert.match(await page.locator('#connection-error').textContent(),/确认清除/);
  await page.locator('#clear-token-confirm').check();await page.evaluate(()=>replies.push({error:'配置版本冲突',status:409}));
  await page.locator('#save-connections').click();await idle(page);assert.equal(await page.locator('#clear-token-confirm').isChecked(),true);assert.equal(await dirty(page),true);
  await page.selectOption('#token-action','replace');await page.locator('#connection-token').fill('synthetic-secret-token');
  await page.evaluate(()=>replies.push({error:'模拟写入故障',status:500}));await page.locator('#save-connections').click();await idle(page);
  assert.equal(await page.locator('#connection-token').inputValue(),'synthetic-secret-token');assert.equal(await dirty(page),true);
  const saved=fixture();saved.backupId='fixture-backup';saved.restartRequired=true;
  await page.evaluate(data=>replies.push(data),saved);await page.locator('#save-connections').click();await idle(page);
  assert.equal(await page.locator('#connection-token').inputValue(),'');assert.equal(await page.locator('#token-action').inputValue(),'keep');assert.equal(await dirty(page),false);
  assert.match(await page.locator('#connection-save-state').textContent(),/待重启.*备份已创建/);
});

test('busy requests freeze the form and block logout; departure warns only for drafts or pending work',async t=>{
  const page=await harness(t);const unload=()=>page.evaluate(()=>{const e=new Event('beforeunload',{cancelable:true});window.dispatchEvent(e);return e.defaultPrevented;});
  assert.equal(await unload(),false);await page.locator('#connection-owners').fill('23456789');assert.equal(await unload(),true);
  await page.locator('#connection-owners').fill('12345678');await page.locator('#reload-connections').click();await page.waitForFunction(()=>pending.length===1);
  assert.equal(await unload(),true);assert.equal(await page.locator('#connection-fields').evaluate(el=>el.disabled),true);
  for(const id of ['connection-url','connection-select','connection-owners','token-action','save-connections'])assert.equal(await page.locator('#'+id).isDisabled(),true);
  await page.locator('#logout').click();assert.equal(await page.evaluate(()=>requests.length),1);
  await page.evaluate(data=>pending[0]({ok:true,status:200,json:async()=>data}),fixture());await idle(page);assert.equal(await unload(),false);
  await page.selectOption('#token-action','replace');await page.locator('#connection-token').fill('logout-secret');
  await page.evaluate(()=>replies.push({}));await confirmAction(page,true,()=>page.locator('#logout').click());await page.waitForFunction(()=>!authenticated);
  for(const id of ['connection-url','connection-token','connection-owners'])assert.equal(await page.locator('#'+id).inputValue(),'');
  assert.equal(await page.locator('#connection-select option').count(),0);assert.equal(await dirty(page),false);assert.equal(await unload(),false);
});

test('expired sessions and late reads/writes cannot refill, expire or unlock a newly logged-in form',async t=>{
  const page=await harness(t);await page.selectOption('#token-action','replace');await page.locator('#connection-token').fill('old-secret');
  await page.locator('#save-connections').click();await page.waitForFunction(()=>pending.length===1);
  await showSession(page,'new-session');await page.locator('#reload-connections').click();await page.waitForFunction(()=>pending.length===2);
  const old=fixture();old.superUsers=['99887766'];
  await page.evaluate(data=>pending[0]({ok:true,status:200,json:async()=>data}),old);await page.waitForTimeout(30);
  assert.equal(await page.evaluate(()=>connectionBusy),true);assert.equal(await page.locator('#connection-owners').inputValue(),'12345678');assert.equal(await page.locator('#connection-token').inputValue(),'');
  await showSession(page,'third-session');await page.locator('#reload-connections').click();await page.waitForFunction(()=>pending.length===3);
  await page.evaluate(()=>pending[1]({ok:false,status:401,json:async()=>({error:'旧会话过期'})}));await page.waitForTimeout(30);
  assert.equal(await page.evaluate(()=>authenticated),true);assert.equal(await page.locator('#reload-connections').isDisabled(),true);
  await page.evaluate(data=>pending[2]({ok:true,status:200,json:async()=>data}),fixture());await idle(page);
  await page.locator('#connection-owners').fill('23456789');await page.evaluate(()=>replies.push({error:'会话过期',status:401}));await page.locator('#save-connections').click();await page.waitForFunction(()=>!authenticated);
  assert.equal(await page.evaluate(()=>connectionSettings),null);assert.equal(await page.locator('#connection-owners').inputValue(),'');assert.equal(await page.locator('#connection-error').isVisible(),false);
});
