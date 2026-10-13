// Synthetic evidence only: no live program, credentials, services or QQ.
import assert from 'node:assert/strict';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import crypto from 'node:crypto';
import {spawnSync} from 'node:child_process';
import test from 'node:test';
import {fileURLToPath} from 'node:url';

const checker=fileURLToPath(new URL('./check-local-admin-delivery.mjs',import.meta.url));
const sha=file=>crypto.createHash('sha256').update(fs.readFileSync(file)).digest('hex').toUpperCase();
const json=file=>JSON.parse(fs.readFileSync(file,'utf8'));
function fixture() {
  const root=fs.mkdtempSync(path.join(os.tmpdir(),'delivery-evidence-'));
  const start=Date.now()-60000;
  const write=(name,value,time=start)=>{
    const file=path.join(root,name);fs.mkdirSync(path.dirname(file),{recursive:true});
    fs.writeFileSync(file,typeof value==='string'||Buffer.isBuffer(value)?value:JSON.stringify(value));fs.utimesSync(file,new Date(time),new Date(time));return file;
  };
  const run=args=>spawnSync(process.execPath,[checker,...args],{cwd:root,encoding:'utf8',windowsHide:true});
  const runtime=(prefix,offset=0)=>{
    const base='.local/'+prefix;
    const bytes=Buffer.from('synthetic-program-'+prefix);
    write('.local/bin/zerobot-'+prefix+'-tested.exe',bytes,start+offset+10000);
    write('.local/bin/zerobot.exe',bytes,start+offset+10000);
    write(base+'-go-tests.txt',['internal/localadmin','internal/basiccommands','plugin/llm','plugin/wordcount'].map(name=>'ok github.com/FloatTech/ZeroBot-Plugin/'+name+' 1.0s').join('\n'));
    write(base+'-vet.txt','');
    write(base+'-build.txt','all modules verified\nBuild ready: zerobot-'+prefix+'-tested.exe\n');
    write(base+'-smoke.txt','ARTIFACTS=/fixture/.local/smoke-admin/run-'+(offset+1)+'\n');
    write('.local/smoke-admin/run-'+(offset+1)+'/result.json',{passed:true,fixturePassed:true,completedAt:new Date(start+offset+20000).toISOString()});
    write(base+'-deployment.json',{programMatchesStaged:true,startupConfigUnchanged:true,controlDatabaseUnchanged:true,dailyCredentialUnchanged:true,napcatProcessPreserved:true,qqLoggedIn:true,onebotReadOnlyAPIsPassed:true,singleZeroBotInstance:true,processIdentityVerified:true,adminHTTPStatus:200,startupErrorDetected:false,completedAt:new Date(start+offset+30000).toISOString()});
  };
  write('internal/localadmin/assets/index.html','<main>fixture-marker</main>');
  write('docs/local-admin/README.md','[guide](guide.md)');write('docs/local-admin/guide.md','Fixture guide');
  write('scripts/smoke-local-admin.go','{"/", "text/html; charset=utf-8", `fixture-marker`}');
  const stats='# tests 2\n# pass 2\n# fail 0\n# cancelled 0\n# skipped 0\n# todo 0\n';
  write('.local/ui-ui-tests.txt',stats);write('.local/ui-focused-ui.txt',stats);
  write('.local/ui-layout-result.json',{passed:true,offline:true,networkRequestsAllowed:false,measurements:[{}]});
  runtime('ui');
  const initial=run(['--evidence-prefix','ui']);assert.equal(initial.status,0,initial.stdout+initial.stderr);
  const baselineFile=path.join(root,'.local/ui-ui-baseline.json');
  const baselineHash=sha(baselineFile);
  runtime('backend',10000);
  return {root,start,write,run,baselineFile,baselineHash,args:['--evidence-prefix','backend','--ui-evidence-prefix','ui'],result:()=>json(path.join(root,'.local/current-delivery-evidence.json'))};
}
test('single prefix creates baseline; split evidence preserves it and identifies original UI time',()=>{
  const f=fixture(),result=f.run(f.args);assert.equal(result.status,0,result.stdout+result.stderr);
  const report=f.result();assert.equal(report.passed,true);assert.equal(report.uiEvidencePrefix,'ui');assert.equal(report.evidencePrefix,'backend');assert.equal(report.reusedUIEvidence,true);assert.equal(report.testsExecuted,false);assert.equal(report.serviceAccessed,false);
  assert.notEqual(report.sourceTestTimes.uiDeployment,report.sourceTestTimes.deployment);assert.equal(sha(f.baselineFile),f.baselineHash);
});
test('changed asset content with original modification time refuses reuse',()=>{
  const f=fixture();f.write('internal/localadmin/assets/index.html','<main>fixture-marker changed</main>');
  assert.equal(f.run(f.args).status,1);assert.ok(f.result().failures.includes('全部界面资源保持一致'));assert.equal(sha(f.baselineFile),f.baselineHash);
});
test('added and removed resources refuse reuse',()=>{
  for(const action of ['add','remove']) {
    const f=fixture();if(action==='add')f.write('internal/localadmin/assets/added.js','fixture');else fs.unlinkSync(path.join(f.root,'internal/localadmin/assets/index.html'));
    assert.equal(f.run(f.args).status,1);assert.ok(f.result().failures.includes('全部界面资源保持一致'));
  }
});
test('changed passing UI evidence refuses reuse',()=>{
  const f=fixture();f.write('.local/ui-ui-tests.txt','# tests 3\n# pass 3\n# fail 0\n# cancelled 0\n# skipped 0\n# todo 0\n');
  assert.equal(f.run(f.args).status,1);assert.ok(f.result().failures.includes('原界面证据摘要一致'));
});
test('missing, duplicate, wrong-prefix and future baselines refuse reuse',()=>{
  for(const action of ['missing','duplicate','wrong-prefix','future']) {
    const f=fixture(),baseline=json(f.baselineFile);
    if(action==='missing')fs.unlinkSync(f.baselineFile);
    else {
      if(action==='duplicate')baseline.assetHashes.push(baseline.assetHashes[0]);
      if(action==='wrong-prefix')baseline.evidencePrefix='other';
      if(action==='future')baseline.sourceTestTimes.deployment=new Date(Date.now()+3600000).toISOString();
      fs.writeFileSync(f.baselineFile,JSON.stringify(baseline));
    }
    assert.equal(f.run(f.args).status,1,action);assert.equal(f.result().passed,false);
    if(action==='missing')assert.equal(fs.existsSync(f.baselineFile),false);
  }
});
test('invalid argument forms leave result and baseline untouched',()=>{
  const f=fixture(),resultFile=path.join(f.root,'.local/current-delivery-evidence.json'),resultHash=sha(resultFile);
  for(const args of [[],['--evidence-prefix','../unsafe'],['--evidence-prefix','backend','--ui-evidence-prefix','../unsafe'],['--evidence-prefix','backend','--unknown','ui'],['--evidence-prefix','backend','--evidence-prefix','ui'],['--evidence-prefix','backend','--ui-evidence-prefix']]) {
    assert.equal(f.run(args).status,2);assert.equal(sha(resultFile),resultHash);assert.equal(sha(f.baselineFile),f.baselineHash);
  }
});
test('uncaptured Write-Host readiness requires matching real upgrade snapshot hash',()=>{
  const f=fixture();f.write('.local/backend-build.txt','go version go1.26.1 windows/amd64\nall modules verified\n');
  const staged=path.join(f.root,'.local/bin/zerobot-backend-tested.exe');f.write('.local/backend-before.json',{stagedHash:sha(staged)});
  assert.equal(f.run(f.args).status,0);assert.equal(f.result().buildReadyMarkerRecorded,false);
  f.write('.local/backend-before.json',{stagedHash:'A'.repeat(64)});assert.equal(f.run(f.args).status,1);assert.ok(f.result().failures.includes('构建就绪'));
});
test('split evidence still rejects wrong live program and failed backend tests',()=>{
  for(const action of ['wrong-program','failed-tests']) {
    const f=fixture();if(action==='wrong-program')f.write('.local/bin/zerobot.exe','different');else fs.appendFileSync(path.join(f.root,'.local/backend-go-tests.txt'),'\nFAIL fixture\n');
    assert.equal(f.run(f.args).status,1);assert.equal(f.result().passed,false);
  }
});
