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
const add=await fs.readFile('internal/localadmin/assets/connections-add.js','utf8');
const browser=await chromium.launch({channel:'chrome',headless:true});
after(()=>browser.close());
function fixture(){return {available:true,revision:'a'.repeat(64),connections:[{kind:'ws',index:0,url:'ws://127.0.0.1:16700',tokenConfigured:true},{kind:'wss',index:0,url:'ws://127.0.0.1:16701',tokenConfigured:false}],superUsers:['12345678'],runningSuperUsers:['12345678'],runningConnections:[{kind:'ws',index:0,url:'ws://127.0.0.1:16700',tokenConfigured:true}],restartRequired:false};}
async function harness(t){
  const context=await browser.newContext();t.after(()=>context.close());await context.route('**/*',r=>r.abort());
  const page=await context.newPage();await page.setContent(html);await page.addScriptTag({content:app});await page.addScriptTag({content:add});
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

const summary=page=>page.locator('#connection-draft-field-list').textContent();
const labels=(page,part)=>page.locator('[data-connection-draft-part="'+part+'"] li').allTextContents();
const assertClean=async page=>{assert.equal(await page.locator('#connection-draft-fields').isHidden(),true);assert.equal(await summary(page),'');};

test('fixed field names combine Token controls, omit values, and revert or discard only the edited part',async t=>{
  const page=await harness(t);await assertClean(page);
  await page.locator('#connection-url').fill('ws://synthetic-secret-address/path');await page.selectOption('#token-action','replace');await page.locator('#connection-token').fill('<img src=x onerror=secret>');await page.locator('#connection-owners').fill('99887766');
  await page.locator('#connection-add-url').fill('ws://synthetic-new-secret');await page.locator('#connection-add-token').fill('synthetic-new-token');await page.selectOption('#connection-add-kind','wss');
  assert.deepEqual(await labels(page,'connection'),['WebSocket 地址','Token']);assert.deepEqual(await labels(page,'owners'),['主人 QQ 账号']);assert.deepEqual(await labels(page,'addition'),['连接方向','WebSocket 地址','Token']);
  assert.match(await summary(page),/正向连接 1.*未保存字段：2 项/);assert.match(await summary(page),/尚未保存的新连接.*未保存字段：3 项/);
  const html=await page.locator('#connection-draft-fields').innerHTML();assert.doesNotMatch(html,/synthetic-secret|synthetic-new|99887766|onerror|a{64}/);assert.equal(await page.locator('#connection-draft-fields img').count(),0);
  await page.selectOption('#token-action','clear');await page.locator('#clear-token-confirm').check();assert.deepEqual(await labels(page,'connection'),['WebSocket 地址','Token']);
  await page.selectOption('#token-action','keep');assert.deepEqual(await labels(page,'connection'),['WebSocket 地址']);await page.locator('#connection-url').fill(fixture().connections[0].url);assert.equal(await page.locator('[data-connection-draft-part="connection"]').count(),0);
  await confirmAction(page,true,()=>page.locator('#discard-connection-owners').click());assert.equal(await page.locator('[data-connection-draft-part="owners"]').count(),0);assert.deepEqual(await labels(page,'addition'),['连接方向','WebSocket 地址','Token']);
  await page.locator('#connection-add-discard').click();await page.locator('#connection-add-confirm-accept').click();await assertClean(page);assert.equal(await page.evaluate(()=>requests.length),0);
});

test('actual loaded connection sets the target; owner and addition survive switching; unknown metadata never appears',async t=>{
  const page=await harness(t);await page.locator('#connection-owners').fill('99887766');await page.locator('#connection-add-url').fill('ws://new-synthetic');await page.selectOption('#connection-select','1');await page.locator('#connection-url').fill('ws://reverse-synthetic');
  await page.evaluate(()=>{document.getElementById('connection-select').value='0';renderConnectionDrafts();});assert.match(await summary(page),/反向监听 1/);assert.doesNotMatch(await summary(page),/正向连接 1/);
  await confirmAction(page,true,()=>page.locator('#discard-current-connection').click());assert.deepEqual(await labels(page,'owners'),['主人 QQ 账号']);assert.deepEqual(await labels(page,'addition'),['WebSocket 地址']);
  await page.locator('#connection-url').fill('ws://reverse-synthetic');await confirmAction(page,false,()=>page.selectOption('#connection-select','0'));assert.match(await summary(page),/反向监听 1/);
  await confirmAction(page,true,()=>page.selectOption('#connection-select','0'));assert.equal(await page.locator('[data-connection-draft-part="connection"]').count(),0);
  await page.evaluate(()=>{connectionSettings.connections[0].kind='secret-kind';connectionSettings.connections[0].index='secret-index';document.getElementById('connection-url').value='ws://changed';renderConnectionDrafts();});assert.match(await summary(page),/目标暂不可核对/);assert.doesNotMatch(await summary(page),/secret-kind|secret-index/);assert.equal(await page.evaluate(()=>requests.length),0);
});

test('save resets only existing fields, busy and failed reload or conflict retain summaries; successful reload clears all',async t=>{
  const page=await harness(t);await page.locator('#connection-url').fill('ws://127.0.0.1:16750');await page.locator('#connection-owners').fill('98765432');await page.locator('#connection-add-token').fill('synthetic-unsaved-add');
  await page.locator('#save-connections').click();await page.waitForFunction(()=>pending.length===1);assert.match(await page.locator('#connection-draft-field-state').textContent(),/请求处理中/);assert.deepEqual(await labels(page,'connection'),['WebSocket 地址']);
  const saved=fixture();saved.connections[0].url='ws://127.0.0.1:16750';saved.superUsers=['98765432'];saved.restartRequired=true;saved.revision='b'.repeat(64);saved.backupId='synthetic';
  await page.evaluate(data=>pending[0]({ok:true,status:200,json:async()=>data}),saved);await idle(page);assert.equal(await page.locator('[data-connection-draft-part="connection"]').count(),0);assert.equal(await page.locator('[data-connection-draft-part="owners"]').count(),0);assert.deepEqual(await labels(page,'addition'),['Token']);assert.match(await page.locator('#connection-save-state').textContent(),/待重启/);
  await page.locator('#connection-url').fill('ws://127.0.0.1:16751');await page.evaluate(()=>replies.push({error:'conflict',status:409}));await page.locator('#save-connections').click();await idle(page);assert.deepEqual(await labels(page,'connection'),['WebSocket 地址']);
  const preserved=await summary(page);await page.evaluate(()=>replies.push({error:'read failed',status:500}));await confirmAction(page,true,()=>page.locator('#reload-connections').click());await idle(page);assert.equal(await summary(page),preserved);
  await page.evaluate(data=>replies.push(data),saved);await confirmAction(page,true,()=>page.locator('#reload-connections').click());await idle(page);await assertClean(page);assert.equal(await page.evaluate(()=>requests.length),4);
});

test('addition direction and no-Token count once; failed addition retains draft and successful addition clears it',async t=>{
  const page=await harness(t);await page.locator('#connection-add-token').fill('synthetic-token');await page.locator('#connection-add-without-token').check();assert.equal(await page.locator('#connection-add-token').inputValue(),'');assert.deepEqual(await labels(page,'addition'),['Token']);
  await page.locator('#connection-add-without-token').uncheck();await assertClean(page);await page.selectOption('#connection-add-kind','wss');assert.deepEqual(await labels(page,'addition'),['连接方向']);await page.locator('#connection-add-url').fill('ws://127.0.0.1:16755');await page.locator('#connection-add-token').fill('synthetic-token');
  const submit=async()=>{await page.locator('#connection-add-save').click();assert.match(await page.locator('#connection-add-confirm-message').textContent(),/新增本机反向监听/);await page.locator('#connection-add-confirm-accept').click();await idle(page);};
  await page.evaluate(()=>replies.push({error:'synthetic failed',status:409}));await submit();assert.deepEqual(await labels(page,'addition'),['连接方向','WebSocket 地址','Token']);
  const saved=fixture();saved.connections.push({kind:'wss',index:1,url:'ws://127.0.0.1:16755',tokenConfigured:true});saved.revision='c'.repeat(64);saved.backupId='synthetic-added';saved.restartRequired=true;await page.evaluate(data=>replies.push(data),saved);await submit();await assertClean(page);assert.equal(await page.evaluate(()=>connectionIndex),2);assert.equal(await page.locator('#connection-add-token').inputValue(),'');assert.equal(await page.evaluate(()=>requests.length),2);
});

test('navigation preserves drafts while reset, logout, expiry, unavailable and uneditable initial state hide and clear summaries',async t=>{
  const page=await harness(t);await page.locator('#connection-owners').fill('98765432');const preserved=await summary(page);await page.locator('[data-view="overview"]').click();await page.locator('[data-view="connections"]').click();assert.equal(await summary(page),preserved);
  await page.evaluate(()=>resetConnections());await assertClean(page);
  const unavailable=fixture();unavailable.available=false;await page.evaluate(data=>{renderConnectionSettings(data);document.getElementById('connection-owners').value='98765432';renderConnectionDrafts();},unavailable);await assertClean(page);
  const empty=fixture();empty.connections=[];await page.evaluate(data=>{renderConnectionSettings(data);connectionControls(false);},empty);await page.locator('#connection-owners').fill('98765432');assert.deepEqual(await labels(page,'owners'),['主人 QQ 账号']);assert.equal(await page.locator('[data-connection-draft-part="connection"]').count(),0);
  await showSession(page,'logout-session');await page.locator('#connection-owners').fill('98765432');await page.evaluate(()=>replies.push({}));await confirmAction(page,true,()=>page.locator('#logout').click());await page.waitForFunction(()=>!authenticated);await assertClean(page);
  await showSession(page,'expired-session');await page.locator('#connection-owners').fill('98765432');await page.evaluate(()=>replies.push({error:'expired',status:401}));await page.locator('#save-connections').click();await page.waitForFunction(()=>!authenticated);await assertClean(page);
});

test('late old successful or expired responses cannot refill or unlock a fresh session summary',async t=>{
  const page=await harness(t);await page.locator('#connection-url').fill('ws://127.0.0.1:16750');await page.locator('#save-connections').click();await page.waitForFunction(()=>pending.length===1);
  await showSession(page,'new-session');await page.locator('#connection-owners').fill('98765432');await page.locator('#connection-add-url').fill('ws://new-synthetic');const preserved=await summary(page);
  await page.evaluate(data=>pending[0]({ok:true,status:200,json:async()=>data}),fixture());await page.waitForTimeout(40);assert.equal(await summary(page),preserved);assert.equal(await page.evaluate(()=>authenticated),true);
  await confirmAction(page,true,()=>page.locator('#reload-connections').click());await page.waitForFunction(()=>pending.length===2);await showSession(page,'third-session');await page.selectOption('#connection-add-kind','wss');const current=await summary(page);await confirmAction(page,true,()=>page.locator('#reload-connections').click());await page.waitForFunction(()=>pending.length===3);
  await page.evaluate(()=>pending[1]({ok:false,status:401,json:async()=>({error:'old expiry'})}));await page.waitForTimeout(40);assert.equal(await summary(page),current);assert.equal(await page.evaluate(()=>connectionBusy),true);assert.equal(await page.evaluate(()=>authenticated),true);assert.equal(await page.locator('#connection-add-discard').isDisabled(),true);
  await page.evaluate(data=>pending[2]({ok:true,status:200,json:async()=>data}),fixture());await idle(page);await assertClean(page);
});

