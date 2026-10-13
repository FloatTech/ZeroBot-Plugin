import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs/promises';
import path from 'node:path';
import {fileURLToPath} from 'node:url';
import {spawnSync} from 'node:child_process';
import {checkCombinedObservation,combineObservation,combinedExitCode} from './check-combined-observation.mjs';
import {stamp,validateResource} from './check-observation-evidence.mjs';
const root=await fs.mkdtemp(path.join(path.dirname(fileURLToPath(import.meta.url)),'../.local/combined-observation-tests-'));
const clone=v=>structuredClone(v),modelKinds=['gpt','chat','agent','vision'],replyKinds=['gpt','chat','agent'];
const modelZero={started:0,active:0,completed:0,failed:0,incomplete:0,averageMilliseconds:0,maxMilliseconds:0};
const replyZero={started:0,active:0,completed:0,failed:0,noReply:0,withReply:0,missingIngress:0,replyAttempts:0,replyAccepted:0,replyUnconfirmed:0,averageFirstReplyMilliseconds:0,maxFirstReplyMilliseconds:0,averageLastReplyMilliseconds:0,maxLastReplyMilliseconds:0};
function fixture(){
 const process={processId:12345,processStartedAt:'2026-10-13T00:00:00.000000100Z',programSHA256:'A'.repeat(64)};
 const modelSnapshot=(checkedAt,done=false)=>{const row=done?{...modelZero,started:7,completed:7,averageMilliseconds:150,maxMilliseconds:200}:modelZero;return {startedAt:'2026-10-13T00:00:01Z',checkedAt,...row,categories:modelKinds.map(kind=>({kind,...(kind==='gpt'?row:modelZero)})),process:clone(process)}};
 const delta={started:7,completed:7,failed:0,incomplete:0,activeAtStart:0,activeAtEnd:0};
 const model={format:'zerobot-model-observation',version:2,exportedAt:'2026-10-13T00:20:00Z',scope:{kind:'model-http-exchanges',latencyScope:'process-cumulative',businessAcceptancePassed:false},before:modelSnapshot('2026-10-13T00:01:00Z'),after:modelSnapshot('2026-10-13T00:10:00Z',true),window:{durationSeconds:540,...delta,categories:modelKinds.map(kind=>({kind,...(kind==='gpt'?delta:{started:0,completed:0,failed:0,incomplete:0,activeAtStart:0,activeAtEnd:0})}))}};
 const replyReport=(checkedAt,done=false)=>{const row=done?{...replyZero,started:3,completed:3,withReply:3,replyAttempts:6,replyAccepted:6,averageFirstReplyMilliseconds:200,maxFirstReplyMilliseconds:400,averageLastReplyMilliseconds:300,maxLastReplyMilliseconds:500}:replyZero;return {format:'zerobot-reply-statistics',version:1,exportedAt:'2026-10-13T00:20:00Z',scope:{kind:'ai-reply-tasks',latencyScope:'process-cumulative',timingBoundary:'bot-matcher-entry-to-onebot-send-receipt',businessAcceptancePassed:false},snapshot:{startedAt:'2026-10-13T00:00:01Z',checkedAt,scope:'bot-matcher-entry-to-onebot-send-receipt',...row,categories:replyKinds.map(kind=>({kind,...(kind==='gpt'?row:replyZero)})),process:clone(process)}}};
 const samples=Array.from({length:181},(_,i)=>({at:new Date(Date.parse('2026-10-13T00:00:02Z')+i*5000).toISOString(),atSeconds:i*5,cpuSeconds:10+i/100,privateBytes:1000000+(i?8192:0),workingSetBytes:2000000,threadCount:12,handleCount:200}));
 const resource={startedAt:'2026-10-13T00:00:01.999Z',completedAt:'2026-10-13T00:15:02.100Z',...process,requestedDurationSeconds:900,intervalSeconds:5,logicalCPUs:20,thresholds:{cpuAverageCapacityPercent:20,privateGrowthMiB:64},completed:true,passed:true,identityStable:true,programUnchanged:true,stopReason:'duration-complete',realQQMessagesSent:false,modelCalls:0,businessLoadMeasured:false,businessAcceptancePassed:false,messageContentsSaved:false,servicesStartedOrStopped:false,noNewErrorLines:true,logCounts:{appendedLines:0,warningLines:0,errorLines:0,fatalLines:0,panicLines:0,otherLines:0,bytesRead:0,unreadBytes:0,partialLines:0,logTruncated:false,coverageComplete:true}};
 return {model,before:replyReport('2026-10-13T00:01:30.123456789Z'),after:replyReport('2026-10-13T00:09:00.987654321Z',true),resource,samples};
}
let serial=0;
async function save(f){const dir=path.join(root,String(++serial));await fs.mkdir(dir);const summary=validateResource({...f.resource,resources:{}},f.samples.map(s=>({...s,ns:stamp(s.at).ns})));f.resource.resources=summary.resources;f.resource.passed=summary.resourceChecksPassed;for(const key of ['model','before','after'])await fs.writeFile(path.join(dir,key+'.json'),JSON.stringify(f[key]));await fs.writeFile(path.join(dir,'result.json'),JSON.stringify(f.resource));await fs.writeFile(path.join(dir,'samples.jsonl'),f.samples.map(s=>JSON.stringify(s)).join('\n')+'\n');return {resource:dir,model:path.join(dir,'model.json'),before:path.join(dir,'before.json'),after:path.join(dir,'after.json')};}
async function check(f){return checkCombinedObservation(await save(f));}

test('overlapping windows preserve separate model tasks and receipts without equating counts or inferring business',async()=>{
 const r=await check(fixture());assert.equal(r.evidenceValid,true);assert.equal(r.resourceChecksPassed,true);assert.equal(r.sameProcessVerified,true);assert.equal(r.windowsOverlap,true);assert.equal(r.compatible,true);assert.equal(combinedExitCode(r),0);assert.equal(r.model.model.window.completed,7);assert.equal(r.reply.window.completed,3);assert.equal(r.reply.window.replyAccepted,6);assert.equal(r.businessAcceptancePassed,false);assert.equal(r.overlap.startedAt,'2026-10-13T00:01:30.123456789Z');assert.equal(r.overlap.endedAt,'2026-10-13T00:09:00.987654321Z');assert.equal(r.overlap.durationNanoseconds,'450864197532');assert.equal(r.overlap.durationSeconds,450.864197532);assert.equal(r.overlap.averageMilliseconds,undefined);
});

test('individually compatible windows that are disjoint or only touch cannot form a joint observation',async t=>{
 for(const [name,start,end,kind] of [['disjoint','2026-10-13T00:10:00.000000001Z','2026-10-13T00:11:00Z','none'],['touch','2026-10-13T00:10:00Z','2026-10-13T00:11:00Z','boundary-only']])await t.test(name,async()=>{const f=fixture();f.before.snapshot.checkedAt=start;f.after.snapshot.checkedAt=end;const r=await check(f);assert.equal(r.model.compatible,true);assert.equal(r.reply.compatible,true);assert.equal(r.evidenceValid,true);assert.equal(r.sameProcessVerified,true);assert.equal(r.windowsOverlap,false);assert.equal(r.compatible,false);assert.equal(r.overlap.kind,kind);assert.equal(combinedExitCode(r),2);});
});

test('positive one nanosecond intersection remains exact',async()=>{
 const f=fixture();f.before.snapshot.checkedAt='2026-10-13T00:09:59.999999999Z';f.after.snapshot.checkedAt='2026-10-13T00:11:00Z';const r=await check(f);assert.equal(combinedExitCode(r),0);assert.equal(r.overlap.durationNanoseconds,'1');assert.equal(r.overlap.durationSeconds,1e-9);
});

test('resource replacements between two successful readings are rejected by both source digests',async t=>{
 const base=await check(fixture());for(const key of ['resourceSHA256','samplesSHA256'])await t.test(key,()=>{const model=clone(base.model),reply=clone(base.reply);reply.sources[key]='B'.repeat(64);const r=combineObservation(model,reply);assert.equal(r.evidenceValid,false);assert.equal(r.sameProcessVerified,false);assert.equal(r.compatible,false);assert.equal(combinedExitCode(r),1);assert.ok(r.issues.includes('sources:resource-readings-differ-or-missing'))});
});

test('valid but missing identity or failed resource conditions remain insufficient',async t=>{
 const cases=[['legacy-model',f=>{f.model.version=1;delete f.model.before.process;delete f.model.after.process}],['no-reply-identity',f=>{delete f.before.snapshot.process;delete f.after.snapshot.process}],['wrong-model-id',f=>{f.model.before.process.processId++;f.model.after.process.processId++}],['wrong-reply-id',f=>{f.before.snapshot.process.processId++;f.after.snapshot.process.processId++}],['short',f=>f.resource.requestedDurationSeconds=10],['threshold',f=>f.resource.thresholds.privateGrowthMiB=0.001],['resource-restarted',f=>{f.resource.identityStable=false;f.resource.completed=false;f.resource.stopReason='process-identity-changed'}],['resource-clock',f=>{for(let i=1;i<f.samples.length;i++)f.samples[i].at=new Date(Date.parse(f.samples[i].at)+1500).toISOString();f.resource.completedAt='2026-10-13T00:15:04Z'}],['partial-model',f=>{f.model.before.checkedAt='2026-10-13T00:00:01.500Z';f.model.window.durationSeconds=598.5}],['partial-reply',f=>f.before.snapshot.checkedAt='2026-10-13T00:00:01.500Z']];
 for(const [name,change] of cases)await t.test(name,async()=>{const f=fixture();change(f);const r=await check(f);assert.equal(r.evidenceValid,true);assert.equal(r.compatible,false);assert.equal(combinedExitCode(r),2);assert.equal(r.businessAcceptancePassed,false);});
});

test('invalid model reply pair and corrupt resource files fail with no secret contents',async t=>{
 const cases=[['model-count',f=>f.model.after.started++],['reply-count',f=>f.after.snapshot.started++],['reply-restart',f=>f.after.snapshot.startedAt='2026-10-13T00:00:01.000000001Z'],['model-restart',f=>f.model.after.startedAt='2026-10-13T00:00:01.000000001Z']];for(const [name,change] of cases)await t.test(name,async()=>{const f=fixture();change(f);const r=await check(f);assert.equal(r.evidenceValid,false);assert.equal(r.compatible,false);assert.equal(combinedExitCode(r),1)});
 const input=await save(fixture());await fs.writeFile(input.after,'PRIVATE-BAD-JSON');const r=await checkCombinedObservation(input);assert.equal(combinedExitCode(r),1);assert.ok(!JSON.stringify(r).includes('PRIVATE'));await fs.writeFile(path.join(input.resource,'result.json'),'PRIVATE-BAD-RESOURCE');assert.equal(combinedExitCode(await checkCombinedObservation(input)),1);
});

test('private extra fields are discarded across every source',async()=>{
 const f=fixture();f.resource.path='PRIVATE-RESOURCE-PATH';f.model.before.process.path='PRIVATE-PROGRAM';f.model.key='PRIVATE-MODEL-KEY';f.before.snapshot.account='PRIVATE-ACCOUNT';f.after.snapshot.categories[0].text='PRIVATE-MESSAGE';f.samples[0].text='PRIVATE-LOG';const r=await check(f);assert.equal(combinedExitCode(r),0);assert.ok(!JSON.stringify(r).includes('PRIVATE'));
});

test('CLI new output and argument errors preserve all original files and existing output',async()=>{
 const input=await save(fixture()),output=path.join(input.resource,'combined.json'),tool=fileURLToPath(new URL('check-combined-observation.mjs',import.meta.url)),run=args=>spawnSync(process.execPath,[tool,...args],{encoding:'utf8'});const args=Object.entries(input).flatMap(([k,v])=>['--'+k,v]);const files=[input.model,input.before,input.after,path.join(input.resource,'result.json'),path.join(input.resource,'samples.jsonl')],original=await Promise.all(files.map(p=>fs.readFile(p)));const first=run([...args,'--output',output]);assert.equal(first.status,0,first.stderr);assert.equal(JSON.parse(first.stdout).compatible,true);assert.equal(JSON.parse(await fs.readFile(output,'utf8')).format,'zerobot-combined-observation-check');const originalOutput=await fs.readFile(output);assert.equal(run([...args,'--output',output]).status,1);assert.deepEqual(await fs.readFile(output),originalOutput);assert.equal(run([...args,'--output',input.model]).status,1);for(let i=0;i<files.length;i++)assert.deepEqual(await fs.readFile(files[i]),original[i]);
 for(const bad of [[],['--resource','PRIVATE'],['--unknown','PRIVATE'],[...args,'--model',input.model],[...args,'--output']]){const r=run(bad);assert.equal(r.status,1);assert.ok(!r.stderr.includes('PRIVATE'))}
 const f=fixture();f.before.snapshot.checkedAt='2026-10-13T00:10:00Z';f.after.snapshot.checkedAt='2026-10-13T00:11:00Z';const disjoint=await save(f),other=Object.entries(disjoint).flatMap(([k,v])=>['--'+k,v]);assert.equal(run(other).status,2);
});
