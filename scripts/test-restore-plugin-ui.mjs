// Offline browser with synthetic state only; every network request is blocked.
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
const backups=await fs.readFile('internal/localadmin/assets/backups.js','utf8');
const browser=await chromium.launch({channel:'chrome',headless:true});
after(()=>browser.close());
async function fixture(t) {
  const context=await browser.newContext();t.after(()=>context.close());
  await context.route('**/*',route=>route.abort());
  const p=await context.newPage();await p.setContent(html);await p.addScriptTag({content:app});await p.addScriptTag({content:backups});
  await p.evaluate(()=>{
    authenticated=true;csrf='synthetic-session';window.showOverview=()=>{};
    window.readMode='ready';window.pendingReads=[];window.pendingSaves=[];window.requests=[];window.restored=false;window.empty=false;
    window.list=enabled=>({controlAvailable:true,plugins:empty?[]:[{name:'chrev',brief:'翻转字符',usage:'/chrev',defaultEnabledAtStartup:true,state:{effectiveEnabled:enabled,defaultEnabled:true,globalMode:'inherit',groupMode:enabled?'enabled':'disabled',source:'group',revision:(enabled?'a':'b').repeat(64)}}],response:{globalEnabled:enabled,groupEnabled:false,effectiveEnabled:enabled,source:enabled?'global':'none'}});
    const initial=list(true);plugins=initial.plugins;pluginResponse=initial.response;pluginResponseLoaded=true;pluginControlAvailable=true;pluginGroup='101';pluginCatalogReadAt=new Date().toISOString();
    document.getElementById('auth').hidden=true;document.getElementById('workspace').hidden=false;
    document.getElementById('plugin-scope').value='group';document.getElementById('plugin-group').value='101';
    const entry={id:'synthetic-control-backup',scope:'control-db',valid:true,integrity:true,version:1,size:512};
    window.api=async(url,body)=>{
      requests.push({url,body});
      if(url==='/api/backups')return {backups:[entry],scopes:{'control-db':true}};
      if(url==='/api/backups/preview')return {backup:entry,changed:true,revision:'preview-version',token:'synthetic-ticket',expiresAt:Math.floor(Date.now()/1000)+600,tables:1,rows:1,changedTables:1,warnings:[]};
      if(url==='/api/backups/restore'){restored=true;return {changed:true,restartRequired:true};}
      if(url.startsWith('/api/plugins?')) {
        if(readMode==='pending')return new Promise((resolve,reject)=>pendingReads.push({resolve,reject}));
        if(readMode==='fail')throw Error('合成读取失败');
        return list(!restored);
      }
      if(url==='/api/plugin-switch')return new Promise(resolve=>pendingSaves.push(resolve));
      throw Error('Unexpected fixture API: '+url);
    };
    renderPluginScope();renderPlugins();
  });
  await p.locator('[data-view="plugins"]').click();
  return p;
}
async function restore(p) {
  await p.locator('[data-view="backups"]').click();
  await p.locator('#backup-list button').click();await p.locator('#restore-confirm').check();
  await p.locator('#restore-backup').click();await p.waitForFunction(()=>!backupBusy&&restored);
}
const readRequests=p=>p.evaluate(()=>requests.filter(r=>r.url.startsWith('/api/plugins?')).length);
const idle=p=>p.waitForFunction(()=>!pluginScopeBusy&&!pluginCatalogLoading);

test('returning after restore automatically reloads the same scope and response, preserving filters; concurrent navigation does not duplicate reads',async t=>{
  const p=await fixture(t);await p.locator('#plugin-search').fill('chrev');await p.selectOption('#plugin-sort','name');
  await restore(p);await p.evaluate(()=>readMode='pending');await p.locator('[data-view="plugins"]').click();
  await p.waitForFunction(()=>pendingReads.length===1);assert.match(await p.locator('#plugin-empty').textContent(),/正在读取/);assert.equal(await p.locator('#reload-plugins').isDisabled(),true);
  await p.locator('[data-view="overview"]').click();await p.locator('[data-view="plugins"]').click();assert.equal(await readRequests(p),1);
  await p.evaluate(()=>pendingReads[0].resolve(list(false)));await idle(p);
  assert.deepEqual(await p.locator('#plugin-list h2').allTextContents(),['chrev']);assert.match(await p.locator('#plugin-list').textContent(),/当前生效：禁用/);
  assert.match(await p.locator('#plugin-response-state').textContent(),/全局响应：关闭/);assert.equal(await p.locator('#plugin-search').inputValue(),'chrev');assert.equal(await p.locator('#plugin-sort').inputValue(),'name');
  assert.equal(await p.locator('#plugin-group').inputValue(),'101');assert.equal(await p.evaluate(()=>pluginDrafts.size),0);
  assert.equal(await p.evaluate(()=>requests.filter(r=>r.url==='/api/plugin-switch').length),0);
});

test('failed automatic read reports failure, manual retry recovers, and a successful empty list is not fetched on every visit',async t=>{
  const p=await fixture(t);await restore(p);await p.evaluate(()=>readMode='fail');await p.locator('[data-view="plugins"]').click();await idle(p);
  assert.match(await p.locator('#plugin-empty').textContent(),/未能读取.*刷新/);assert.match(await p.locator('#plugin-error').textContent(),/合成读取失败/);assert.equal(await p.locator('#reload-plugins').isEnabled(),true);
  await p.evaluate(()=>{readMode='ready';empty=true;});await p.locator('#reload-plugins').click();await idle(p);assert.match(await p.locator('#plugin-empty').textContent(),/当前程序未注册/);
  await p.locator('[data-view="overview"]').click();await p.locator('[data-view="plugins"]').click();assert.equal(await readRequests(p),2);
});

test('a read started before restore cannot replace the restored list or freeze its controls',async t=>{
  const p=await fixture(t);await p.evaluate(()=>readMode='pending');await p.locator('#reload-plugins').click();await p.waitForFunction(()=>pendingReads.length===1);
  await restore(p);await p.evaluate(()=>readMode='ready');await p.locator('[data-view="plugins"]').click();await idle(p);
  await p.evaluate(()=>pendingReads[0].resolve(list(true)));await p.waitForTimeout(20);
  assert.match(await p.locator('#plugin-list').textContent(),/当前生效：禁用/);assert.equal(await p.locator('#reload-plugins').isEnabled(),true);assert.equal(await readRequests(p),2);
});

test('a save response from before restore cannot repopulate drafts/status or release the new read lock',async t=>{
  const p=await fixture(t);await p.locator('#plugin-list select').selectOption('disabled');await p.locator('#plugin-list .plugin-controls button').click();await p.waitForFunction(()=>pendingSaves.length===1);
  await restore(p);await p.evaluate(()=>readMode='pending');await p.locator('[data-view="plugins"]').click();await p.waitForFunction(()=>pendingReads.length===1);
  await p.evaluate(()=>pendingSaves[0]({state:list(false).plugins[0].state}));await p.waitForTimeout(20);
  assert.equal(await p.locator('#reload-plugins').isDisabled(),true);assert.equal(await p.locator('#plugin-save-state').textContent(),'');assert.equal(await p.evaluate(()=>pluginDrafts.size),0);
  await p.evaluate(()=>pendingReads[0].resolve(list(false)));await idle(p);assert.match(await p.locator('#plugin-list').textContent(),/当前生效：禁用/);
});
