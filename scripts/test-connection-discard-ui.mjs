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

test('cancel preserves both drafts; local restores are independent and retain restart and running-state notices',async t=>{
  const page=await harness(t),data=fixture();data.restartRequired=true;data.superUsers=['12345678','23456789'];
  await page.evaluate(data=>{renderConnectionSettings(data);connectionControls(false);},data);
  const savedState=await page.locator('#connection-save-state').textContent(),running=await page.locator('#connection-running').textContent(),runningOwners=await page.locator('#connection-running-owners').textContent();
  assert.equal(await page.locator('#discard-current-connection').isDisabled(),true);assert.equal(await page.locator('#discard-connection-owners').isDisabled(),true);
  await page.locator('#connection-url').fill('ws://127.0.0.1:16799');await page.selectOption('#token-action','replace');await page.locator('#connection-token').fill('synthetic-private-token');await page.locator('#connection-owners').fill('34567890');
  await confirmAction(page,false,()=>page.locator('#discard-current-connection').click());await confirmAction(page,false,()=>page.locator('#discard-connection-owners').click());
  assert.equal(await page.locator('#connection-url').inputValue(),'ws://127.0.0.1:16799');assert.equal(await page.locator('#connection-token').inputValue(),'synthetic-private-token');assert.equal(await page.locator('#connection-owners').inputValue(),'34567890');
  await confirmAction(page,true,()=>page.locator('#discard-connection-owners').click());assert.equal(await page.locator('#connection-owners').inputValue(),'12345678, 23456789');assert.equal(await page.locator('#connection-token').inputValue(),'synthetic-private-token');assert.equal(await page.locator('#token-action').inputValue(),'replace');
  const message=await confirmAction(page,true,()=>page.locator('#discard-current-connection').click());assert.match(message,/正向连接 1/);assert.doesNotMatch(message,/synthetic-private-token|16799/);
  assert.equal(await page.locator('#connection-url').inputValue(),data.connections[0].url);assert.equal(await page.locator('#connection-token').inputValue(),'');assert.equal(await page.locator('#token-action').inputValue(),'keep');assert.equal(await page.locator('#connection-token').isDisabled(),true);assert.equal(await page.locator('#connection-token').evaluate(el=>el.required),false);assert.equal(await page.locator('#clear-token-confirm').isChecked(),false);assert.equal(await page.locator('#clear-token-label').isHidden(),true);
  assert.equal(await page.locator('#connection-save-state').textContent(),savedState);assert.match(savedState,/待重启/);assert.equal(await page.locator('#connection-running').textContent(),running);assert.equal(await page.locator('#connection-running-owners').textContent(),runningOwners);assert.equal(await dirty(page),false);assert.equal(await page.evaluate(()=>requests.length),0);
});

test('owner discard preserves token-clear confirmation and connection discard restores the loaded reverse target only',async t=>{
  const page=await harness(t);await page.selectOption('#connection-select','1');await page.locator('#connection-owners').fill('34567890');await page.selectOption('#token-action','clear');await page.locator('#clear-token-confirm').check();
  await confirmAction(page,true,()=>page.locator('#discard-connection-owners').click());assert.equal(await page.locator('#token-action').inputValue(),'clear');assert.equal(await page.locator('#clear-token-confirm').isChecked(),true);assert.equal(await page.locator('#clear-token-label').isVisible(),true);
  await page.locator('#connection-owners').fill('45678901');await page.locator('#connection-url').fill('ws://127.0.0.1:16788');
  await page.evaluate(()=>{document.getElementById('connection-select').value='0';});
  const message=await confirmAction(page,true,()=>page.locator('#discard-current-connection').click());assert.match(message,/反向监听 1/);
  assert.equal(await page.locator('#connection-select').inputValue(),'1');assert.equal(await page.evaluate(()=>connectionIndex),1);assert.equal(await page.locator('#connection-url').inputValue(),fixture().connections[1].url);assert.equal(await page.locator('#connection-owners').inputValue(),'45678901');assert.equal(await page.locator('#token-action').inputValue(),'keep');assert.equal(await page.locator('#clear-token-confirm').isChecked(),false);
  assert.deepEqual(await page.evaluate(()=>connectionSettings.connections),fixture().connections);assert.equal(await page.evaluate(()=>requests.length),0);
});

test('last successful saved values remain the baseline after conflict, without undoing saved settings or notices',async t=>{
  const page=await harness(t);await page.locator('#connection-url').fill('ws://127.0.0.1:16750');await page.locator('#connection-owners').fill('56789012');
  const saved=fixture();saved.connections[0].url='ws://127.0.0.1:16750';saved.superUsers=['56789012'];saved.revision='b'.repeat(64);saved.restartRequired=true;saved.backupId='synthetic-backup';
  await page.evaluate(data=>replies.push(data),saved);await page.locator('#save-connections').click();await idle(page);
  const notice=await page.locator('#connection-save-state').textContent();assert.match(notice,/待重启.*备份已创建/);
  await page.locator('#connection-url').fill('ws://127.0.0.1:16751');await page.selectOption('#token-action','replace');await page.locator('#connection-token').fill('synthetic-conflict-token');await page.locator('#connection-owners').fill('67890123');
  await page.evaluate(()=>replies.push({error:'版本冲突',status:409}));await page.locator('#save-connections').click();await idle(page);assert.equal(await page.locator('#connection-token').inputValue(),'synthetic-conflict-token');
  await confirmAction(page,true,()=>page.locator('#discard-current-connection').click());assert.equal(await page.locator('#connection-url').inputValue(),'ws://127.0.0.1:16750');assert.equal(await page.locator('#connection-owners').inputValue(),'67890123');assert.equal(await page.locator('#connection-token').inputValue(),'');
  await confirmAction(page,true,()=>page.locator('#discard-connection-owners').click());assert.equal(await page.locator('#connection-owners').inputValue(),'56789012');assert.equal(await page.locator('#connection-save-state').textContent(),notice);assert.equal(await page.evaluate(()=>connectionSettings.revision),'b'.repeat(64));assert.equal(await dirty(page),false);assert.equal(await page.evaluate(()=>requests.length),2);
});

test('unavailable and empty-connection configurations disable connection undo, while owner undo remains available when editable',async t=>{
  const page=await harness(t),unavailable=fixture();unavailable.available=false;
  await page.evaluate(data=>{renderConnectionSettings(data);connectionControls(false);document.getElementById('connection-owners').value='23456789';renderConnectionDrafts();discardConnectionDraft('owners');},unavailable);
  assert.equal(await page.locator('#discard-current-connection').isDisabled(),true);assert.equal(await page.locator('#discard-connection-owners').isDisabled(),true);assert.equal(await page.evaluate(()=>requests.length),0);
  const empty=fixture();empty.connections=[];empty.runningConnections=[];await page.evaluate(data=>{renderConnectionSettings(data);connectionControls(false);},empty);
  await page.locator('#connection-owners').fill('');assert.equal(await page.locator('#connection-owners').evaluate(el=>el.checkValidity()),false);assert.equal(await page.locator('#discard-current-connection').isDisabled(),true);assert.equal(await page.locator('#discard-connection-owners').isDisabled(),false);
  await confirmAction(page,true,()=>page.locator('#discard-connection-owners').click());assert.equal(await page.locator('#connection-owners').inputValue(),'12345678');assert.equal(await dirty(page),false);assert.equal(await page.evaluate(()=>requests.length),0);
});

test('busy requests, newer loaded settings and sessions prevent stale undo; logout and expiry clear all local inputs',async t=>{
  const page=await harness(t);await page.locator('#connection-url').fill('ws://127.0.0.1:16760');await page.locator('#connection-owners').fill('34567890');await page.locator('#save-connections').click();await page.waitForFunction(()=>pending.length===1);
  for(const id of ['discard-current-connection','discard-connection-owners'])assert.equal(await page.locator('#'+id).isDisabled(),true);
  await page.evaluate(()=>{discardConnectionDraft('connection');discardConnectionDraft('owners');});assert.equal(await page.locator('#connection-url').inputValue(),'ws://127.0.0.1:16760');assert.equal(await page.locator('#connection-owners').inputValue(),'34567890');
  await page.evaluate(data=>pending[0]({ok:true,status:200,json:async()=>data}),fixture());await idle(page);
  await page.locator('#connection-owners').fill('45678901');
  await page.locator('#discard-connection-owners').click();
  await page.evaluate(data=>{renderConnectionSettings(data);connectionControls(false);document.getElementById('connection-owners').value='87654321';renderConnectionDrafts();},fixture());
  await page.locator('#connection-draft-confirm-accept').click();
  assert.equal(await page.locator('#connection-owners').inputValue(),'87654321');
  await page.locator('#connection-url').fill('ws://127.0.0.1:16770');
  await page.locator('#discard-current-connection').click();
  await showSession(page,'new-session');
  await page.locator('#connection-draft-confirmation').waitFor({state:'hidden'});
  await page.locator('#connection-url').fill('ws://127.0.0.1:16780');await page.selectOption('#token-action','replace');await page.locator('#connection-token').fill('synthetic-new-session-token');
  assert.equal(await page.locator('#connection-url').inputValue(),'ws://127.0.0.1:16780');assert.equal(await page.locator('#connection-token').inputValue(),'synthetic-new-session-token');
  await confirmAction(page,true,()=>page.locator('#reload-connections').click());await page.waitForFunction(()=>pending.length===2);
  await showSession(page,'third-session');await page.locator('#connection-owners').fill('98765432');await page.evaluate(()=>pending[1]({ok:false,status:401,json:async()=>({error:'old expiry'})}));await page.waitForTimeout(30);
  assert.equal(await page.evaluate(()=>authenticated),true);assert.equal(await page.locator('#connection-owners').inputValue(),'98765432');assert.equal(await page.locator('#discard-connection-owners').isDisabled(),false);
  await page.evaluate(()=>replies.push({error:'expired',status:401}));await page.locator('#save-connections').click();await page.waitForFunction(()=>!authenticated);assert.equal(await page.locator('#connection-token').inputValue(),'');assert.equal(await page.locator('#connection-discard-status').textContent(),'');assert.equal(await page.locator('#discard-current-connection').isDisabled(),true);assert.equal(await page.locator('#discard-connection-owners').isDisabled(),true);
  await showSession(page,'last-session');await page.locator('#connection-owners').fill('98765432');await page.evaluate(()=>replies.push({}));await confirmAction(page,true,()=>page.locator('#logout').click());await page.waitForFunction(()=>!authenticated);assert.equal(await page.locator('#discard-connection-owners').isDisabled(),true);assert.equal(await page.locator('#connection-owners').inputValue(),'');
});
