import test from 'node:test';
import assert from 'node:assert/strict';
import {mkdtemp,writeFile,readFile,open,mkdir,symlink} from 'node:fs/promises';
import path from 'node:path';
import {fileURLToPath} from 'node:url';
import {spawnSync} from 'node:child_process';
import {checkEvidence,checkResourceEvidence,evidenceExitCode,stamp,validateModel,validateResource,writeEvidence} from './check-observation-evidence.mjs';

const repo=path.resolve(path.dirname(fileURLToPath(import.meta.url)),'..');
const root=await mkdtemp(path.join(repo,'.local','observation-evidence-tests-'));
const clone=value=>structuredClone(value),kinds=['gpt','chat','agent','vision'];
const zero={started:0,active:0,completed:0,failed:0,incomplete:0,averageMilliseconds:0,maxMilliseconds:0};
function fixture(){
  const samples=Array.from({length:181},(_,i)=>({at:new Date(Date.parse('2026-10-13T00:00:02Z')+i*5000).toISOString(),atSeconds:i*5,cpuSeconds:10+i/100,privateBytes:1000000+(i?8192:0),workingSetBytes:2000000,threadCount:12,handleCount:200}));
  const meta={startedAt:'2026-10-13T08:00:01.999+08:00',completedAt:'2026-10-13T08:15:02.100+08:00',processId:12345,processStartedAt:'2026-10-13T00:00:00Z',programSHA256:'A'.repeat(64),requestedDurationSeconds:900,intervalSeconds:5,logicalCPUs:20,thresholds:{cpuAverageCapacityPercent:20,privateGrowthMiB:64},completed:true,passed:true,identityStable:true,programUnchanged:true,stopReason:'duration-complete',realQQMessagesSent:false,modelCalls:0,businessLoadMeasured:false,businessAcceptancePassed:false,messageContentsSaved:false,servicesStartedOrStopped:false,noNewErrorLines:true,logCounts:{appendedLines:0,warningLines:0,errorLines:0,fatalLines:0,panicLines:0,otherLines:0,bytesRead:0,unreadBytes:0,partialLines:0,logTruncated:false,coverageComplete:true},resources:{sampleCount:181,sampleSpanSeconds:900,maxSampleGapSeconds:5,cpuAveragePercent:1.8/900/20*100,cpuPeakSamplePercent:0.01/5/20*100,privateFirstBytes:1000000,privateMaxBytes:1008192,privateLastBytes:1008192,privateGrowthMiB:0.0078125,workingSetMaxBytes:2000000,threadMax:12,handleMax:200,metricsValid:true,completeWindow:true,cadencePassed:true,resourceThresholdsPassed:true,fullTiming:true}};
  const a={...zero,started:2,active:2},b={...zero,started:6,active:1,completed:5,averageMilliseconds:3,maxMilliseconds:5};
  const snapshot=(checkedAt,gpt)=>({startedAt:'2026-10-13T00:00:01.000000001Z',checkedAt,...gpt,categories:kinds.map(kind=>({kind,...(kind==='gpt'?gpt:zero)}))});
  const delta={started:4,completed:5,failed:0,incomplete:0,activeAtStart:2,activeAtEnd:1};
  const model={format:'zerobot-model-observation',version:1,exportedAt:'2026-10-13T00:10:01Z',scope:{kind:'model-http-exchanges',latencyScope:'process-cumulative',businessAcceptancePassed:false},before:snapshot('2026-10-13T00:01:00Z',a),after:snapshot('2026-10-13T00:10:00Z',b),window:{durationSeconds:540,...delta,categories:kinds.map(kind=>({kind,...(kind==='gpt'?delta:{started:0,completed:0,failed:0,incomplete:0,activeAtStart:0,activeAtEnd:0})}))}};
  return {meta,samples,model};
}
function summarize(f){
  const result=validateResource(f.meta,f.samples.map(r=>({...r,ns:stamp(r.at).ns})));
  f.meta.resources=result.resources;f.meta.passed=result.resourceChecksPassed;
}
let serial=0;
async function save(f){
  const dir=path.join(root,String(++serial));await mkdir(dir);
  await writeFile(path.join(dir,'result.json'),JSON.stringify(f.meta));
  await writeFile(path.join(dir,'samples.jsonl'),f.samples.map(r=>JSON.stringify(r)).join('\n')+'\n');
  const model=path.join(dir,'model.json');await writeFile(model,JSON.stringify(f.model));
  return {resource:dir,model};
}
async function check(f){return checkEvidence(await save(f));}

test('resource only independently recomputes a formal window without inventing model association',async()=>{
 const input=await save(fixture());await writeFile(input.model,'NOT A MODEL REPORT');
 const r=await checkResourceEvidence({resource:input.resource});assert.equal(r.mode,'resource-only');assert.equal(r.evidenceValid,true);assert.equal(r.resourceObservationPassed,true);assert.equal(r.resourceChecksPassed,true);assert.equal(evidenceExitCode(r),0);assert.equal(r.resource.resources.privateGrowthMiB,0.0078125);assert.equal(r.compatible,false);assert.equal(r.sameProcessVerified,false);assert.equal(r.businessAcceptancePassed,false);assert.equal(r.temporalCoverage,'unavailable');assert.equal(r.model,undefined);assert.equal(r.sources.modelSHA256,undefined);assert.ok(r.issues.includes('association:model-report-not-provided'));
});

test('resource only preserves diagnostic outcomes for insufficient timing and failed checks',async t=>{
 const changes=[['short',f=>f.meta.requestedDurationSeconds=10],['stopped',f=>{f.meta.completed=false;f.meta.stopReason='stop-request'}],['cadence',f=>f.samples.splice(1,10)],['threshold',f=>f.meta.thresholds.privateGrowthMiB=0.001],['identity',f=>f.meta.identityStable=false],['program',f=>f.meta.programUnchanged=false],['log-error',f=>{f.meta.logCounts={...f.meta.logCounts,appendedLines:1,errorLines:1,bytesRead:50};f.meta.noNewErrorLines=false}],['log-coverage',f=>{f.meta.logCounts.unreadBytes=1;f.meta.logCounts.coverageComplete=false}]];
 for(const [name,change] of changes)await t.test(name,async()=>{const f=fixture();change(f);summarize(f);const r=await checkResourceEvidence(await save(f));assert.equal(r.evidenceValid,true);assert.equal(r.resourceObservationPassed,false);assert.equal(evidenceExitCode(r),2);assert.equal(r.businessAcceptancePassed,false);});
});

test('resource only rejects mismatched summaries and corrupt or missing samples without rewriting',async t=>{
 for(const [name,change] of [['rounded',f=>f.meta.resources.privateGrowthMiB=0],['secret-flag',f=>f.meta.messageContentsSaved=true]])await t.test(name,async()=>{const f=fixture();change(f);const input=await save(f),original=await readFile(path.join(input.resource,'result.json'));const r=await checkResourceEvidence(input);assert.equal(r.evidenceValid,false);assert.equal(r.resourceObservationPassed,false);assert.equal(evidenceExitCode(r),1);assert.deepEqual(await readFile(path.join(input.resource,'result.json')),original)});
 const input=await save(fixture());await writeFile(path.join(input.resource,'samples.jsonl'),'PRIVATE-BAD-SAMPLE');const r=await checkResourceEvidence(input);assert.equal(evidenceExitCode(r),1);assert.ok(!JSON.stringify(r).includes('PRIVATE-BAD-SAMPLE'));assert.equal(evidenceExitCode(await checkResourceEvidence({resource:path.join(root,'missing')})),1);
});

test('resource only flags inconsistent wall clock even when monotonic resource checks pass',async()=>{
 const f=fixture();for(let i=1;i<f.samples.length;i++)f.samples[i].at=new Date(Date.parse(f.samples[i].at)+1500).toISOString();f.meta.completedAt='2026-10-13T00:15:04Z';summarize(f);const r=await checkResourceEvidence(await save(f));assert.equal(r.evidenceValid,true);assert.equal(r.resourceChecksPassed,true);assert.equal(r.resource.clockConsistent,false);assert.equal(r.resourceObservationPassed,false);assert.equal(evidenceExitCode(r),2);assert.ok(r.issues.includes('resource:sample-clock-drift'));
});

test('resource only copies fixed fields and excludes secrets and paths',async()=>{
 const f=fixture();f.meta.secret='PRIVATE-CREDENTIAL';f.meta.path='PRIVATE-PATH';f.meta.logCounts.message='PRIVATE-MESSAGE';f.samples[0].text='PRIVATE-CONTENT';const r=await checkResourceEvidence(await save(f));assert.equal(r.resourceObservationPassed,true);assert.ok(!JSON.stringify(r).includes('PRIVATE-'));assert.equal(r.resource.processId,f.meta.processId);assert.equal(r.resource.programSHA256,f.meta.programSHA256);
});

test('resource only CLI exits and exclusive output preserve original files',async()=>{
 const input=await save(fixture()),output=path.join(input.resource,'resource-checked.json'),tool=path.join(repo,'scripts','check-observation-evidence.mjs');const run=args=>spawnSync(process.execPath,[tool,...args],{encoding:'utf8'});const args=['--resource',input.resource,'--output',output],r=run(args);assert.equal(r.status,0,r.stderr);assert.equal(JSON.parse(r.stdout).mode,'resource-only');assert.equal(JSON.parse(await readFile(output,'utf8')).resourceObservationPassed,true);assert.equal(run(args).status,1);const original=await readFile(path.join(input.resource,'result.json'));assert.equal(run(['--resource',input.resource,'--output',path.join(input.resource,'result.json')]).status,1);assert.deepEqual(await readFile(path.join(input.resource,'result.json')),original);const f=fixture();f.meta.requestedDurationSeconds=10;summarize(f);const short=await save(f);assert.equal(run(['--resource',short.resource]).status,2);
});

test('compatible full window, fractional growth and crossing active requests remain unverified',async()=>{
  const result=await check(fixture());assert.equal(result.evidenceValid,true);assert.equal(result.compatible,true);assert.equal(result.temporalCoverage,'inside');assert.equal(evidenceExitCode(result),0);
  assert.equal(result.resource.resources.privateGrowthMiB,0.0078125);assert.equal(result.model.window.completed,5);assert.equal(result.model.window.started,4);
  assert.equal(result.sameProcessVerified,false);assert.equal(result.businessAcceptancePassed,false);assert.equal(result.model.window.maxMilliseconds,undefined);
});
test('nanosecond identity and offset/calendar validation',()=>{
  assert.equal(stamp('2026-10-13T08:00:00.123456789+08:00').ns,stamp('2026-10-13T00:00:00.123456789Z').ns);
  assert.equal(stamp('2026-10-13T00:00:00.123456788Z').ns+1n,stamp('2026-10-13T00:00:00.123456789Z').ns);
  for(const value of ['2026-02-30T00:00:00Z','2026-10-13T24:00:00Z','2026-10-13T00:00:60Z','2026-10-13T00:00:00+14:01','2026-10-13T00:00:00+08:60','2026-10-13T00:00:00.1234567890Z','2026-10-13'])assert.throws(()=>stamp(value));
  assert.equal(stamp('2024-02-29T00:00:00-03:30').text,'2024-02-29T03:30:00.000000000Z');
});
test('model format, scope, counts, categories, latency and exported window reject corruption',async t=>{
  const cases=[['format',f=>f.model.version=3],['scope',f=>f.model.scope.businessAcceptancePassed=true],['unsafe-count',f=>f.model.after.started=Number.MAX_SAFE_INTEGER+1],['negative',f=>f.model.after.active=-1],['conservation',f=>f.model.after.started++],['duplicate-category',f=>f.model.before.categories[1].kind='gpt'],['category-total',f=>{f.model.before.categories[0].started=1;f.model.before.categories[0].active=1;}],['latency',f=>f.model.after.averageMilliseconds=6],['empty-latency',f=>f.model.before.maxMilliseconds=1],['window',f=>f.model.window.completed++],['window-latency',f=>f.model.window.maxMilliseconds=5],['category-window-latency',f=>f.model.window.categories[0].averageMilliseconds=3],['export-order',f=>f.model.exportedAt='2026-10-13T00:09:00Z']];
  for(const [name,mutate] of cases)await t.test(name,async()=>{const f=fixture();mutate(f);const r=await check(f);assert.equal(r.evidenceValid,false);assert.equal(evidenceExitCode(r),1);assert.equal(r.compatible,false);});
});
test('model restart, reverse time, same precise time changes and maximum rollback',async t=>{
  for(const [name,mutate] of [['restart',f=>f.model.after.startedAt='2026-10-13T00:00:01.000000002Z'],['reverse',f=>f.model.after.checkedAt='2026-10-13T00:00:59.999999999Z'],['same-time',f=>f.model.after.checkedAt=f.model.before.checkedAt],['maximum',f=>{f.model.before.maxMilliseconds=6;f.model.before.completed=1;f.model.before.active=1;f.model.before.categories[0]={kind:'gpt',...Object.fromEntries(Object.keys(zero).map(k=>[k,f.model.before[k]]))};}]])await t.test(name,async()=>{const f=fixture();mutate(f);assert.throws(()=>validateModel(f.model));});
});
test('zero, partial, unrelated and collector-before-process windows cannot be associated',async t=>{
  const cases=[['zero',f=>{f.model.after=clone(f.model.before);f.model.window={durationSeconds:0,started:0,completed:0,failed:0,incomplete:0,activeAtStart:2,activeAtEnd:2,categories:kinds.map(kind=>({kind,started:0,completed:0,failed:0,incomplete:0,activeAtStart:kind==='gpt'?2:0,activeAtEnd:kind==='gpt'?2:0}))};},'zero-window'],['partial',f=>{f.model.before.checkedAt='2026-10-13T00:00:01.500Z';f.model.window.durationSeconds=598.5;},'partial'],['none',f=>{f.model.before.checkedAt='2026-10-13T00:16:00Z';f.model.after.checkedAt='2026-10-13T00:25:00Z';f.model.exportedAt='2026-10-13T00:25:01Z';},'none'],['collector',f=>{f.model.before.startedAt='2026-10-12T23:59:59Z';f.model.after.startedAt=f.model.before.startedAt;},'inside']];
  for(const [name,mutate,coverage] of cases)await t.test(name,async()=>{const f=fixture();mutate(f);const r=await check(f);assert.equal(r.evidenceValid,true);assert.equal(r.temporalCoverage,coverage);assert.equal(r.compatible,false);assert.equal(evidenceExitCode(r),2);});
});
test('original fractional-memory rounding is detected without rewriting inputs',async()=>{
  const f=fixture();f.meta.resources.privateGrowthMiB=0;const input=await save(f),before=await readFile(path.join(input.resource,'result.json'));
  const result=await checkEvidence(input);assert.equal(result.evidenceValid,false);assert.deepEqual(result.resource.summaryMismatches,['privateGrowthMiB']);assert.equal(result.resource.resources.privateGrowthMiB,0.0078125);assert.deepEqual(await readFile(path.join(input.resource,'result.json')),before);
});
test('submillisecond windows stay diagnostic and category ordering is canonical',async()=>{
  const f=fixture();f.model.before.checkedAt='2026-10-13T00:01:00.000000001Z';f.model.after.checkedAt='2026-10-13T00:01:00.000000002Z';f.model.window.durationSeconds=0;
  for(const key of ['before','after','window'])f.model[key].categories.reverse();
  const r=await check(f);assert.equal(r.evidenceValid,true);assert.equal(r.temporalCoverage,'zero-window');assert.equal(r.compatible,false);assert.deepEqual(r.model.window.categories.map(r=>r.kind),kinds);
});
test('rollback with otherwise consistent totals is rejected',()=>{
  const f=fixture();f.model.after={...f.model.after,...zero,started:1,active:1,categories:kinds.map(kind=>({kind,...zero,...(kind==='gpt'?{started:1,active:1}:{})}))};
  assert.throws(()=>validateModel(f.model),/model-counter-rollback/);
});
test('short, early stop, sparse, thresholds, identity, program and log coverage stay separate from validity',async t=>{
  const cases=[['short',f=>{f.meta.requestedDurationSeconds=10;f.samples=f.samples.slice(0,3);}],['stop',f=>{f.meta.completed=false;f.meta.stopReason='stop-request';f.samples=f.samples.slice(0,5);}],['sparse',f=>f.samples=f.samples.filter((_,i)=>i%6===0)],['cpu',f=>f.meta.thresholds.cpuAverageCapacityPercent=0.00001],['memory',f=>f.meta.thresholds.privateGrowthMiB=0.001],['identity',f=>{f.meta.identityStable=false;f.meta.completed=false;f.meta.stopReason='process-identity-changed';}],['program',f=>f.meta.programUnchanged=false],['error',f=>{f.meta.logCounts.errorLines=1;f.meta.logCounts.appendedLines=1;f.meta.logCounts.bytesRead=100;f.meta.noNewErrorLines=false;}],['partial-log',f=>{f.meta.logCounts.partialLines=1;f.meta.logCounts.coverageComplete=false;}],['unread-log',f=>{f.meta.logCounts.unreadBytes=1;f.meta.logCounts.coverageComplete=false;}],['truncated-log',f=>{f.meta.logCounts.logTruncated=true;f.meta.logCounts.coverageComplete=false;}]];
  for(const [name,mutate] of cases)await t.test(name,async()=>{const f=fixture();mutate(f);summarize(f);const r=await check(f);assert.equal(r.evidenceValid,true);assert.equal(r.resourceChecksPassed,false);assert.equal(r.compatible,false);assert.equal(evidenceExitCode(r),2);});
});
test('wall-clock drift invalidates time association without hiding valid resource arithmetic',async()=>{
  const f=fixture();for(let i=1;i<f.samples.length;i++)f.samples[i].at=new Date(Date.parse(f.samples[i].at)+2000).toISOString();f.meta.completedAt='2026-10-13T00:15:05Z';
  const r=await check(f);assert.equal(r.evidenceValid,true);assert.equal(r.resourceChecksPassed,true);assert.equal(r.resource.clockConsistent,false);assert.equal(r.compatible,false);assert.ok(r.issues.includes('association:sample-clock-drift'));
});
test('resource corruption rejects bad parameters, flags, identity, logs, clocks and raw values',async t=>{
  const cases=[['cpu-zero',f=>f.meta.logicalCPUs=0],['interval',f=>f.meta.intervalSeconds=31],['duration',f=>f.meta.requestedDurationSeconds=9],['threshold',f=>f.meta.thresholds.privateGrowthMiB=0],['identity',f=>f.meta.programSHA256='oops'],['flags',f=>f.meta.passed='true'],['stop-reason',f=>f.meta.stopReason='stop-request'],['scope',f=>f.meta.modelCalls=1],['logs',f=>f.meta.logCounts.appendedLines=1],['log-flags',f=>f.meta.logCounts.coverageComplete=false],['cpu-rollback',f=>f.samples[1].cpuSeconds=0],['elapsed-rollback',f=>f.samples[1].atSeconds=0],['clock-rollback',f=>f.samples[1].at=f.samples[0].at],['negative-memory',f=>f.samples[1].privateBytes=-1],['unsafe-bytes',f=>f.samples[1].privateBytes=Number.MAX_SAFE_INTEGER+1],['start-order',f=>f.meta.startedAt=f.meta.completedAt],['end-order',f=>f.meta.completedAt=f.meta.startedAt],['one-sample',f=>f.samples=f.samples.slice(0,1)],['forged-summary',f=>f.meta.resources.cpuAveragePercent=0],['forged-pass',f=>f.meta.passed=false]];
  for(const [name,mutate] of cases)await t.test(name,async()=>{const f=fixture();mutate(f);const r=await check(f);assert.equal(r.evidenceValid,false);assert.equal(r.compatible,false);assert.equal(evidenceExitCode(r),1);});
});
test('unknown fields and file errors never disclose messages, credentials or input paths',async()=>{
  const secret='SHOULD-NOT-LEAK-SECRET';const f=fixture();f.meta.password=secret;f.meta.logCounts.message=secret;f.samples[0].body=secret;f.model.before.account=secret;f.model.api=secret;f.model.window.message=secret;
  const input=await save(f),r=await checkEvidence(input);assert.equal(r.compatible,true);assert.ok(!JSON.stringify(r).includes(secret));assert.ok(!JSON.stringify(r).includes(root));assert.equal(Object.hasOwn(r.resource,'password'),false);
  const failed=await checkEvidence({resource:path.join(root,secret),model:path.join(root,secret)});assert.equal(failed.evidenceValid,false);assert.ok(!JSON.stringify(failed).includes(secret));
});
test('JSON and JSONL bounds, blank lines, BOM and final unterminated line',async t=>{
  for(const [name,replace,expected] of [['huge-json',async input=>writeFile(input.model,' '.repeat(1024*1024+1)),false],['huge-line',async input=>writeFile(path.join(input.resource,'samples.jsonl'),'x'.repeat(4097)),false],['blank-line',async input=>writeFile(path.join(input.resource,'samples.jsonl'),'\n'),false],['invalid-json',async input=>writeFile(input.model,'{"private":"SECRET"'),false],['bom-final-line',async input=>{const b=await readFile(input.model);await writeFile(input.model,Buffer.concat([Buffer.from([239,187,191]),b]));const s=await readFile(path.join(input.resource,'samples.jsonl'),'utf8');await writeFile(path.join(input.resource,'samples.jsonl'),s.trimEnd());},true]])await t.test(name,async()=>{const input=await save(fixture());await replace(input);const r=await checkEvidence(input);assert.equal(r.evidenceValid,expected);});
});
test('exclusive output refuses source, existing file and symbolic-link targets',async()=>{
  const input=await save(fixture()),r=await checkEvidence(input),output=path.join(input.resource,'checked.json');await writeEvidence(output,r);assert.deepEqual(JSON.parse(await readFile(output,'utf8')),r);
  const original=await readFile(input.model);await assert.rejects(writeEvidence(input.model,r));assert.deepEqual(await readFile(input.model),original);await assert.rejects(writeEvidence(output,r));
  const link=path.join(input.resource,'linked.json');
  // File symlinks may be unavailable without Windows developer mode. Existing-file guard above always runs.
  try{await symlink(input.model,link,'file');}catch(e){if(['EPERM','EACCES'].includes(e.code))return;throw e;}
  await assert.rejects(writeEvidence(link,r));assert.deepEqual(await readFile(input.model),original);
});
test('streaming sample count and byte limits are enforced before resource processing',async t=>{
  for(const [name,padding,count,code] of [['count','',100001,'sample-count-limit'],['bytes','x'.repeat(800),80000,'sample-size-limit']])await t.test(name,async()=>{
    const f=fixture(),input=await save(f),file=await open(path.join(input.resource,'samples.jsonl'),'w');
    const line=JSON.stringify({...f.samples[0],ignored:padding})+'\n';
    try{for(let n=0;n<count;n+=1000)await file.writeFile(line.repeat(Math.min(1000,count-n)));}finally{await file.close();}
    const r=await checkEvidence(input);assert.equal(r.evidenceValid,false);assert.ok(r.issues.includes('resource:'+code),r.issues.join(','));
  });
});
test('CLI exit codes, new output and safe diagnostics',async()=>{
  const input=await save(fixture()),output=path.join(input.resource,'cli.json'),tool=path.join(repo,'scripts','check-observation-evidence.mjs');
  const run=args=>spawnSync(process.execPath,[tool,...args],{encoding:'utf8'});
  const args=['--resource',input.resource,'--model',input.model,'--output',output],first=run(args);assert.equal(first.status,0,first.stderr);assert.equal(JSON.parse(first.stdout).compatible,true);assert.equal(run(args).status,1);
  for(const args of [[],['--model','SECRET'],['--resource',input.resource,'--resource',input.resource],['--unknown','SECRET']]){const r=run(args);assert.equal(r.status,1);assert.ok(!r.stderr.includes('SECRET'));}
  const f=fixture();f.meta.programUnchanged=false;summarize(f);const bad=await save(f);assert.equal(run(['--resource',bad.resource,'--model',bad.model]).status,2);
});
function versionTwo(f){
 f.model.version=2;
 const process={processId:f.meta.processId,processStartedAt:stamp(f.meta.processStartedAt).text,programSHA256:f.meta.programSHA256};
 f.model.before.process=clone(process);f.model.after.process=clone(process);return f;
}
test('version two matches exact OS identity with offset and hash-case normalization',async()=>{
 const f=versionTwo(fixture());f.meta.processStartedAt='2026-10-13T08:00:00+08:00';f.model.before.process.programSHA256='a'.repeat(64);
 f.model.before.process.path='PRIVATE-PATH';f.model.after.process.key='PRIVATE-KEY';
 const r=await check(f);assert.equal(r.compatible,true);assert.equal(r.sameProcessVerified,true);assert.equal(r.businessAcceptancePassed,false);assert.deepEqual(r.processComparison,{processIdMatches:true,processStartedAtMatches:true,programSHA256Matches:true});assert.equal(r.model.version,2);assert.ok(!JSON.stringify(r).includes('PRIVATE-'));assert.ok(!r.issues.includes('association:model-report-has-no-process-identity'));
});
test('version two detects PID, one-nanosecond creation time and program mismatch',async t=>{
 for(const [name,change,key] of [['pid',p=>p.processId++,'processIdMatches'],['start',p=>p.processStartedAt='2026-10-13T00:00:00.000000001Z','processStartedAtMatches'],['program',p=>p.programSHA256='B'.repeat(64),'programSHA256Matches']])await t.test(name,async()=>{
   const f=versionTwo(fixture());change(f.model.before.process);f.model.after.process=clone(f.model.before.process);const r=await check(f);assert.equal(r.evidenceValid,true);assert.equal(r.compatible,false);assert.equal(r.sameProcessVerified,false);assert.equal(r.processComparison[key],false);assert.equal(evidenceExitCode(r),2);assert.ok(r.issues.includes('association:process-identity-mismatch'));
 });
});
test('version two identity is required and must be consistent in both snapshots',async t=>{
 for(const [name,change] of [['missing',f=>delete f.model.before.process],['zero-pid',f=>f.model.before.process.processId=0],['bad-hash',f=>f.model.before.process.programSHA256='private'],['bad-time',f=>f.model.before.process.processStartedAt='2026-02-31T00:00:00Z'],['future-start',f=>f.model.before.process.processStartedAt='2026-10-13T00:00:02Z'],['restart',f=>f.model.after.process.processId++],['changed-time',f=>f.model.after.process.processStartedAt='2026-10-13T00:00:00.000000001Z'],['changed-hash',f=>f.model.after.process.programSHA256='B'.repeat(64)]])await t.test(name,async()=>{const f=versionTwo(fixture());change(f);const r=await check(f);assert.equal(r.evidenceValid,false);assert.equal(r.compatible,false);assert.equal(r.sameProcessVerified,false);assert.equal(evidenceExitCode(r),1);});
});
test('matching identity is distinct from timing, resource thresholds and unstable observations',async t=>{
 for(const [name,change,identityMatches] of [['short',f=>f.meta.requestedDurationSeconds=10,true],['threshold',f=>f.meta.thresholds.privateGrowthMiB=0.001,true],['no-overlap',f=>{f.model.before.checkedAt='2026-10-13T00:16:00Z';f.model.after.checkedAt='2026-10-13T00:25:00Z';f.model.exportedAt='2026-10-13T00:25:01Z';},true],['identity-changed',f=>{f.meta.identityStable=false;f.meta.completed=false;f.meta.stopReason='process-identity-changed';},false],['program-changed',f=>f.meta.programUnchanged=false,false]])await t.test(name,async()=>{const f=versionTwo(fixture());change(f);summarize(f);const r=await check(f);assert.equal(r.evidenceValid,true);assert.equal(r.compatible,false);assert.equal(r.sameProcessVerified,identityMatches);assert.equal(r.businessAcceptancePassed,false);assert.equal(evidenceExitCode(r),2);});
});
test('legacy reports cannot claim process identity from unknown extra fields',async()=>{
 const f=fixture();f.model.before.process={processId:f.meta.processId,processStartedAt:f.meta.processStartedAt,programSHA256:f.meta.programSHA256};f.model.after.process=clone(f.model.before.process);
 const r=await check(f);assert.equal(r.evidenceValid,true);assert.equal(r.compatible,true);assert.equal(r.sameProcessVerified,false);assert.equal(r.model.before.process,undefined);assert.ok(r.issues.includes('association:model-report-has-no-process-identity'));
});
test.after(async()=>{await writeFile(path.join(root,'result.json'),JSON.stringify({completed:true,scope:'synthetic-offline-boundary-tests',networkCalls:0,servicesStartedOrStopped:false})+'\n');});
