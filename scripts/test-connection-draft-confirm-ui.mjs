// Offline Chrome with synthetic settings and mocked APIs; never touches QQ or live configuration.
import assert from 'node:assert/strict';
import fs from 'node:fs/promises';
import path from 'node:path';
import {createRequire} from 'node:module';
import test,{after} from 'node:test';
const require=createRequire(import.meta.url);
const modules=process.env.LOCALADMIN_NODE_MODULES||path.join(process.env.USERPROFILE,'.cache/codex-runtimes/codex-primary-runtime/dependencies/node/node_modules');
const {chromium}=require(path.join(modules,'playwright'));
const html=(await fs.readFile('internal/localadmin/assets/index.html','utf8')).replace(/<script\b[^>]*>[\s\S]*?<\/script>/g,'').replace(/<link\b[^>]*>/g,'');
const app=await fs.readFile('internal/localadmin/assets/app.js','utf8'),add=await fs.readFile('internal/localadmin/assets/connections-add.js','utf8');
const browser=await chromium.launch({channel:'chrome',headless:true});after(()=>browser.close());
const settings=()=>({available:true,revision:'synthetic-revision',connections:[{kind:'ws',index:0,url:'ws://localhost:16700',tokenConfigured:true},{kind:'ws',index:1,url:'ws://localhost:16701',tokenConfigured:true}],superUsers:['12345678'],runningSuperUsers:['12345678'],runningConnections:[],restartRequired:false});
async function fixture(t){
 const context=await browser.newContext();t.after(()=>context.close());await context.route('**/*',r=>r.abort());const page=await context.newPage();await page.setContent(html);await page.addScriptTag({content:app});await page.addScriptTag({content:add});
 await page.evaluate(data=>{
  authenticated=true;csrf='synthetic-session';window.showOverview=()=>{};window.showLogs=()=>{};window.showBackups=()=>{};
  document.getElementById('auth').hidden=true;document.getElementById('workspace').hidden=false;
  for(const view of document.querySelectorAll('.content>section'))view.hidden=view.id!=='connections-view';
  window.requests=[];window.pending=[];
  window.fetch=(url,options)=>{requests.push({url,body:options.body?JSON.parse(options.body):null});return new Promise(resolve=>pending.push(resolve));};
  window.confirm=()=>{throw Error('Unexpected native connection confirmation');};renderConnectionSettings(data);connectionControls(false);
 },settings());return page;
}
const visible=p=>p.locator('#connection-draft-confirmation').waitFor({state:'visible'});
const hidden=p=>p.locator('#connection-draft-confirmation').waitFor({state:'hidden'});
const accept=async p=>{await p.locator('#connection-draft-confirm-accept').click();await hidden(p);};
const count=p=>p.evaluate(()=>requests.length);
async function newSession(p){await p.evaluate(data=>{authenticated=false;csrf='';showAuth();authenticated=true;csrf='new-synthetic-session';document.getElementById('auth').hidden=true;document.getElementById('workspace').hidden=false;renderConnectionSettings(data);connectionControls(false);},settings());}

test('cancel focus, Escape and external close retain all drafts and allow reopening without native dialogs',async t=>{
 const p=await fixture(t);await p.locator('#connection-url').fill('ws://localhost:16799');await p.selectOption('#token-action','replace');await p.locator('#connection-token').fill('synthetic-private-token');await p.locator('#connection-owners').fill('23456789');await p.locator('#connection-add-url').fill('ws://localhost:17700');await p.locator('#connection-add-token').fill('synthetic-add-token');
 for(const action of ['discard-current-connection','discard-connection-owners','reload-connections','logout']){
  await p.locator('#'+action).click();await visible(p);assert.equal(await p.evaluate(()=>document.activeElement.id),'connection-draft-confirm-cancel');assert.doesNotMatch(await p.locator('#connection-draft-confirmation').textContent(),/16799|23456789|17700|synthetic-private-token|synthetic-add-token/);
  await p.keyboard.press('Escape');await hidden(p);assert.equal(await count(p),0);assert.equal(await p.locator('#connection-token').inputValue(),'synthetic-private-token');assert.equal(await p.locator('#connection-owners').inputValue(),'23456789');assert.equal(await p.locator('#connection-add-token').inputValue(),'synthetic-add-token');assert.equal(await p.evaluate(()=>document.activeElement.id),action);
 }
 await p.locator('#reload-connections').click();await visible(p);await p.evaluate(()=>document.getElementById('connection-draft-confirmation').close());await hidden(p);await p.waitForFunction(()=>connectionDraftConfirmation===null);assert.equal(await count(p),0);
});

test('changed target draft invalidates confirmation while edits to the other part are preserved',async t=>{
 const p=await fixture(t);await p.locator('#connection-owners').fill('23456789');await p.locator('#discard-connection-owners').click();await visible(p);
 await p.evaluate(()=>document.getElementById('connection-owners').value='34567890');await accept(p);assert.equal(await p.locator('#connection-owners').inputValue(),'34567890');
 await p.locator('#connection-url').fill('ws://localhost:16799');await p.locator('#discard-current-connection').click();await visible(p);await p.evaluate(()=>document.getElementById('connection-url').value='ws://localhost:16798');await accept(p);assert.equal(await p.locator('#connection-url').inputValue(),'ws://localhost:16798');
 await p.locator('#discard-connection-owners').click();await visible(p);await p.evaluate(()=>document.getElementById('connection-url').value='ws://localhost:16797');await accept(p);assert.equal(await p.locator('#connection-owners').inputValue(),'12345678');assert.equal(await p.locator('#connection-url').inputValue(),'ws://localhost:16797');assert.equal(await count(p),0);
});

test('reload and logout cannot submit an accepted confirmation for a changed owner or addition draft',async t=>{
 for(const action of ['reload-connections','logout']){
  const p=await fixture(t);await p.locator('#connection-owners').fill('23456789');await p.locator('#'+action).click();await visible(p);await p.evaluate(()=>document.getElementById('connection-owners').value='34567890');await accept(p);assert.equal(await count(p),0);assert.equal(await p.evaluate(()=>authenticated),true);
  await p.locator('#'+action).click();await visible(p);await p.evaluate(()=>document.getElementById('connection-add-token').value='new-synthetic-add-token');await accept(p);assert.equal(await count(p),0);assert.equal(await p.locator('#connection-add-token').inputValue(),'new-synthetic-add-token');
 }
});

test('reset closes pending confirmation and old selection continuation cannot override a new session',async t=>{
 const p=await fixture(t);await p.locator('#connection-url').fill('ws://localhost:16799');await p.selectOption('#connection-select','1');await visible(p);await newSession(p);await hidden(p);
 await p.selectOption('#connection-select','1');await p.locator('#connection-url').fill('ws://localhost:16798');assert.equal(await p.locator('#connection-select').inputValue(),'1');assert.equal(await p.locator('#connection-url').inputValue(),'ws://localhost:16798');assert.equal(await count(p),0);
});

test('replacement settings reject old acceptance without restoring a stale selection',async t=>{
 const p=await fixture(t);await p.locator('#connection-url').fill('ws://localhost:16799');await p.selectOption('#connection-select','1');await visible(p);
 await p.evaluate(data=>{renderConnectionSettings(data);connectionIndex=1;renderSelectedConnection();document.getElementById('connection-url').value='ws://localhost:16798';},settings());await accept(p);
 assert.equal(await p.locator('#connection-select').inputValue(),'1');assert.equal(await p.locator('#connection-url').inputValue(),'ws://localhost:16798');assert.equal(await count(p),0);
});

test('duplicate and competing addition actions cannot replace or nest a pending confirmation',async t=>{
 const p=await fixture(t);await p.locator('#connection-owners').fill('23456789');await p.locator('#connection-add-url').fill('ws://localhost:17700');await p.locator('#connection-add-token').fill('synthetic-add-token');await p.locator('#discard-connection-owners').click();await visible(p);
 const message=await p.locator('#connection-draft-confirm-message').textContent();await p.evaluate(()=>{loadConnections();discardConnectionDraft('owners');document.getElementById('connection-add-form').dispatchEvent(new Event('submit',{cancelable:true}));});
 assert.equal(await p.locator('#connection-draft-confirm-message').textContent(),message);assert.equal(await p.locator('#connection-add-confirmation').isHidden(),true);assert.equal(await p.evaluate(()=>confirmConnectionAddition('synthetic competing action','confirm')),false);await p.locator('#connection-draft-confirm-cancel').click();await hidden(p);assert.equal(await count(p),0);
 await p.locator('#connection-owners').fill('12345678');await p.locator('#connection-add-save').click();await p.locator('#connection-add-confirmation').waitFor({state:'visible'});await p.evaluate(()=>{document.getElementById('connection-owners').value='23456789';loadConnections();discardConnectionDraft('owners');});assert.equal(await p.locator('#connection-draft-confirmation').isHidden(),true);await p.locator('#connection-add-confirm-cancel').click();assert.equal(await count(p),0);
});

test('delayed close event from a previous modal cannot cancel an immediately reopened confirmation',async t=>{
 const p=await fixture(t);await p.locator('#connection-owners').fill('23456789');await p.locator('#discard-connection-owners').click();await visible(p);
 await p.evaluate(()=>{finishConnectionDraftConfirmation(false);discardConnectionDraft('owners');});await visible(p);await p.waitForTimeout(50);assert.equal(await p.locator('#connection-draft-confirmation').isVisible(),true);await accept(p);assert.equal(await p.locator('#connection-owners').inputValue(),'12345678');assert.equal(await count(p),0);
});

test('a late logout response cannot log out a newer authenticated session',async t=>{
 const p=await fixture(t);await p.locator('#connection-owners').fill('23456789');await p.locator('#logout').click();await visible(p);await accept(p);await p.waitForFunction(()=>pending.length===1);assert.equal((await p.evaluate(()=>requests[0])).url,'/api/logout');
 await newSession(p);await p.locator('#connection-owners').fill('34567890');await p.evaluate(()=>pending[0]({ok:true,status:200,json:async()=>({})}));await p.waitForTimeout(50);assert.equal(await p.evaluate(()=>authenticated),true);assert.equal(await p.locator('#connection-owners').inputValue(),'34567890');
});
