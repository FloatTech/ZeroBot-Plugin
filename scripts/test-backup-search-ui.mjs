// Offline Chrome with mock metadata/APIs only; no actual restore/delete or secrets.
import assert from 'node:assert/strict';
import fs from 'node:fs/promises';
import path from 'node:path';
import {createRequire} from 'node:module';
import test from 'node:test';
const require=createRequire(import.meta.url);
const modules=process.env.LOCALADMIN_NODE_MODULES || path.join(process.env.USERPROFILE,'.cache/codex-runtimes/codex-primary-runtime/dependencies/node/node_modules');
const {chromium}=require(path.join(modules,'playwright'));
const html=(await fs.readFile('internal/localadmin/assets/index.html','utf8')).replace(/<script\b[^>]*>[\s\S]*?<\/script>/g,'').replace(/<link\b[^>]*>/g,'');
const app=await fs.readFile('internal/localadmin/assets/app.js','utf8'),backups=await fs.readFile('internal/localadmin/assets/backups.js','utf8');
test('real backup page combines filters, preserves preview target and filters across app navigation',async t=>{
  const browser=await chromium.launch({channel:'chrome',headless:true});t.after(()=>browser.close());
  const context=await browser.newContext();await context.route('**/*',r=>r.abort());const page=await context.newPage();await page.setContent(html);
  await page.addScriptTag({content:app});await page.addScriptTag({content:backups});
  await page.evaluate(()=>{
    authenticated=true;csrf='fixture';window.showOverview=()=>{};window.showLogs=()=>{};
    document.getElementById('auth').hidden=true;document.getElementById('workspace').hidden=false;
    window.data={backups:[{id:'20261009-control',scope:'control-db',valid:true,integrity:true,size:512,version:1,createdAt:'2026-10-09T00:00:00Z'},{id:'old-startup',scope:'startup-config',valid:true,integrity:false,size:256,version:1,createdAt:'2026-10-08T00:00:00Z'},{id:'<img src=x onerror=bad()>',valid:false}],scopes:{'control-db':true,'startup-config':true},more:false};
    window.requests=[];window.fetch=async(url,options)=>{requests.push({url,body:options.body?JSON.parse(options.body):null});return {ok:true,status:200,json:async()=>url==='/api/backups/preview'?{backup:data.backups[0],changed:true,revision:'revision',token:'ticket',expiresAt:Math.floor(Date.now()/1000)+600,tables:1,rows:1,changedTables:1,warnings:[]}:data};};
  });
  await page.locator('[data-view="backups"]').click();await page.waitForFunction(()=>!backupBusy);
  assert.match(await page.locator('#backup-filter-count').textContent(),/3 \/ 3/);assert.equal(await page.locator('#backup-list img').count(),0);
  await page.selectOption('#backup-state-filter','verified');await page.locator('#backup-list button').click();await page.waitForFunction(()=>!backupBusy);
  assert.equal(await page.locator('#restore-backup-id').textContent(),'备份编号：20261009-control');await page.locator('#restore-confirm').check();
  await page.locator('#backup-search').fill('no-match');assert.equal(await page.locator('#backup-list .backup-card').count(),0);
  assert.equal(await page.locator('#restore-preview').isVisible(),true);assert.equal(await page.locator('#restore-confirm').isChecked(),true);
  assert.equal(await page.evaluate(()=>backupPreview.backup.id),'20261009-control');assert.equal(await page.evaluate(()=>requests.length),2);
  await page.locator('#clear-backup-filter').click();await page.selectOption('#backup-scope-filter','startup-config');await page.selectOption('#backup-state-filter','legacy');
  await page.locator('[data-view="plugins"]').click();await page.locator('[data-view="backups"]').click();await page.waitForFunction(()=>!backupBusy);
  assert.equal(await page.locator('#backup-state-filter').inputValue(),'legacy');assert.equal(await page.locator('#backup-scope-filter').inputValue(),'startup-config');
  assert.deepEqual(await page.locator('#backup-list small').allTextContents(),['old-startup']);assert.equal(await page.locator('#restore-preview').isVisible(),false);
  await page.locator('#clear-backup-filter').click();await page.selectOption('#backup-state-filter','invalid');assert.equal(await page.locator('#backup-list button').isDisabled(),true);
  const bodies=await page.evaluate(()=>requests.filter(r=>r.body));assert.deepEqual(bodies,[{url:'/api/backups/preview',body:{id:'20261009-control'}}]);
});


test('Chrome paging keeps hidden preview target, uses reduced-motion scrolling and places list before cleanup',async t=>{
  const browser=await chromium.launch({channel:'chrome',headless:true});t.after(()=>browser.close());
  const context=await browser.newContext({reducedMotion:'reduce'});await context.route('**/*',r=>r.abort());
  const page=await context.newPage();await page.setContent(html);await page.addScriptTag({content:app});await page.addScriptTag({content:backups});
  await page.evaluate(()=>{
    authenticated=true;csrf='fixture';window.showOverview=()=>{};window.showLogs=()=>{};
    document.getElementById('auth').hidden=true;document.getElementById('workspace').hidden=false;
    window.data={backups:Array.from({length:45},(_,i)=>({id:'fixture-'+i,scope:'control-db',valid:true,integrity:true,size:512,version:1,createdAt:'2026-10-09T00:00:00Z'})),scopes:{'control-db':true},more:false};
    window.requests=[];window.scrollCalls=[];document.getElementById('restore-preview').scrollIntoView=options=>scrollCalls.push(options);
    window.fetch=async(url,options)=>{requests.push({url,body:options.body?JSON.parse(options.body):null});return {ok:true,status:200,json:async()=>url==='/api/backups/preview'?{backup:data.backups[0],changed:true,revision:'revision',token:'ticket',expiresAt:Math.floor(Date.now()/1000)+600,tables:1,rows:1,changedTables:1,warnings:[]}:data};};
  });
  await page.locator('[data-view="backups"]').click();await page.waitForFunction(()=>!backupBusy);
  assert.equal(await page.locator('#backup-list .backup-card').count(),20);assert.ok((await page.locator('#backup-page-status').textContent()).includes('1～20 / 45'));
  assert.equal(await page.evaluate(()=>!!(document.getElementById('backup-list').compareDocumentPosition(document.getElementById('cleanup-form'))&Node.DOCUMENT_POSITION_FOLLOWING)),true);
  await page.locator('#backup-list button').first().click();await page.waitForFunction(()=>!backupBusy);await page.locator('#restore-confirm').check();
  assert.deepEqual(await page.evaluate(()=>scrollCalls),[{block:'start',behavior:'instant'}]);
  await page.locator('#backup-next').click();assert.ok((await page.locator('#backup-page-status').textContent()).includes('21～40 / 45'));
  assert.equal(await page.locator('#restore-backup-id').textContent(),'备份编号：fixture-0');assert.equal(await page.locator('#restore-confirm').isChecked(),true);
  assert.equal((await page.locator('#backup-list small').allTextContents()).includes('fixture-0'),false);
  await page.locator('#backup-next').click();assert.equal(await page.locator('#backup-list .backup-card').count(),5);assert.equal(await page.locator('#backup-next').isDisabled(),true);
  await page.selectOption('#backup-page-size','10');assert.equal(await page.evaluate(()=>backupPage),1);assert.equal(await page.locator('#backup-list .backup-card').count(),10);
  assert.equal(await page.evaluate(()=>requests.length),2);
});
