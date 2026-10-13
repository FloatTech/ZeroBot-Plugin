import test from 'node:test';
import assert from 'node:assert/strict';
import {mkdtemp,mkdir,writeFile,readFile,symlink} from 'node:fs/promises';
import path from 'node:path';
import {fileURLToPath} from 'node:url';
import {spawnSync} from 'node:child_process';
import {validateReplyReport,replyDelta,checkReplyEvidence,replyEvidenceExitCode} from './check-reply-evidence.mjs';
import {stamp,validateResource} from './check-observation-evidence.mjs';
const root=await mkdtemp(path.join(path.dirname(fileURLToPath(import.meta.url)),'../.local/reply-evidence-tests-'));
const zero={started:0,active:0,completed:0,failed:0,noReply:0,withReply:0,missingIngress:0,replyAttempts:0,replyAccepted:0,replyUnconfirmed:0,averageFirstReplyMilliseconds:0,maxFirstReplyMilliseconds:0,averageLastReplyMilliseconds:0,maxLastReplyMilliseconds:0};
const boundary='bot-matcher-entry-to-onebot-send-receipt',clone=v=>structuredClone(v);
function report(row=zero,at='2026-10-13T00:01:00.123456789Z'){
 return {format:'zerobot-reply-statistics',version:1,exportedAt:'2026-10-13T00:20:00Z',scope:{kind:'ai-reply-tasks',latencyScope:'process-cumulative',timingBoundary:boundary,businessAcceptancePassed:false},snapshot:{startedAt:'2026-10-13T00:00:01.123456700Z',checkedAt:at,scope:boundary,...row,categories:['gpt','chat','agent'].map(kind=>({kind,...(kind==='gpt'?row:zero)})),process:{processId:12345,processStartedAt:'2026-10-13T00:00:00.000000100Z',programSHA256:'A'.repeat(64)}}};
}
const done={...zero,started:3,completed:3,withReply:3,replyAttempts:3,replyAccepted:3,averageFirstReplyMilliseconds:100,maxFirstReplyMilliseconds:150,averageLastReplyMilliseconds:200,maxLastReplyMilliseconds:250};
function fixture(){return {before:report(),after:report(done,'2026-10-13T00:10:00.123456789Z')};}
let sequence=0;
async function save(f){const dir=path.join(root,String(++sequence));await mkdir(dir);const before=path.join(dir,'before.json'),after=path.join(dir,'after.json');await writeFile(before,JSON.stringify(f.before));await writeFile(after,JSON.stringify(f.after));return {before,after};}
async function check(f){return checkReplyEvidence(await save(f));}
function resourceFixture(){
 const samples=Array.from({length:181},(_,i)=>({at:new Date(Date.parse('2026-10-13T00:00:02Z')+i*5000).toISOString(),atSeconds:i*5,cpuSeconds:10+i/100,privateBytes:1000000+(i?8192:0),workingSetBytes:2000000,threadCount:12,handleCount:200}));
 const meta={startedAt:'2026-10-13T00:00:01.999Z',completedAt:'2026-10-13T00:15:02.100Z',processId:12345,processStartedAt:'2026-10-13T08:00:00.000000100+08:00',programSHA256:'A'.repeat(64),requestedDurationSeconds:900,intervalSeconds:5,logicalCPUs:20,thresholds:{cpuAverageCapacityPercent:20,privateGrowthMiB:64},completed:true,passed:true,identityStable:true,programUnchanged:true,stopReason:'duration-complete',realQQMessagesSent:false,modelCalls:0,businessLoadMeasured:false,businessAcceptancePassed:false,messageContentsSaved:false,servicesStartedOrStopped:false,noNewErrorLines:true,logCounts:{appendedLines:0,warningLines:0,errorLines:0,fatalLines:0,panicLines:0,otherLines:0,bytesRead:0,unreadBytes:0,partialLines:0,logTruncated:false,coverageComplete:true}};
 return {meta,samples};
}
async function saveResource(f){const dir=path.join(root,'resource-'+(++sequence));await mkdir(dir);const computed=validateResource({...f.meta,resources:{}},f.samples.map(s=>({...s,ns:stamp(s.at).ns})));f.meta.resources=computed.resources;f.meta.passed=computed.resourceChecksPassed;await writeFile(path.join(dir,'result.json'),JSON.stringify(f.meta));await writeFile(path.join(dir,'samples.jsonl'),f.samples.map(s=>JSON.stringify(s)).join('\n')+'\n');return dir;}

test('pair computes exact increments and retains cumulative timing with no inferred business or resource association',async()=>{
 const r=await check(fixture());assert.equal(r.evidenceValid,true);assert.equal(r.replyObservationPassed,true);assert.equal(r.snapshotProcessMatches,true);assert.equal(r.sameProcessVerified,false);assert.equal(r.compatible,false);assert.equal(r.temporalCoverage,'unavailable');assert.equal(r.businessAcceptancePassed,false);assert.equal(replyEvidenceExitCode(r),0);assert.equal(r.window.durationNanoseconds,'540000000000');assert.equal(r.window.durationSeconds,540);assert.equal(r.window.started,3);assert.equal(r.window.completed,3);assert.equal(r.window.replyAccepted,3);assert.equal(r.after.snapshot.averageFirstReplyMilliseconds,100);assert.equal(r.window.averageFirstReplyMilliseconds,undefined);assert.equal(r.window.maxLastReplyMilliseconds,undefined);
});

test('ongoing tasks can finish in a window with more ended reply tasks than new accepted receipts',async()=>{
 const a={...zero,started:2,active:2,missingIngress:1,replyAttempts:1,replyAccepted:1},b={...done,started:4,active:1,completed:1,failed:1,noReply:1,withReply:2,missingIngress:2,replyAttempts:3,replyAccepted:2,replyUnconfirmed:1};const r=await check({before:report(a),after:report(b,'2026-10-13T00:10:00.123456789Z')});assert.equal(replyEvidenceExitCode(r),0);assert.equal(r.window.started,2);assert.equal(r.window.completed+r.window.failed+r.window.noReply,3);assert.equal(r.window.activeAtStart,2);assert.equal(r.window.activeAtEnd,1);assert.equal(r.window.withReply,2);assert.equal(r.window.replyAccepted,1);assert.equal(r.window.replyUnconfirmed,1);
});

test('malformed format boundary counts categories denominators times and process are rejected',async t=>{
 const changes=[['format',v=>v.format='zerobot-model-observation'],['version',v=>v.version=2],['kind',v=>v.scope.kind='OTHER'],['latency-scope',v=>v.scope.latencyScope='window'],['timing-boundary',v=>v.scope.timingBoundary='QQ-client'],['business',v=>v.scope.businessAcceptancePassed=true],['unexpected-window',v=>v.window={}],['scope',v=>v.snapshot.scope='OTHER'],['negative',v=>v.snapshot.started=-1],['unsafe',v=>v.snapshot.started=Number.MAX_SAFE_INTEGER+1],['task-count',v=>v.snapshot.active=1],['receipt-count',v=>v.snapshot.replyAccepted=1],['missing-ingress',v=>v.snapshot.missingIngress=1],['with-reply',v=>v.snapshot.withReply=1],['nonfinite-time',v=>v.snapshot.maxFirstReplyMilliseconds=Infinity],['empty-time',v=>{v.snapshot.maxFirstReplyMilliseconds=1;v.snapshot.maxLastReplyMilliseconds=1}],['category-duplicate',v=>v.snapshot.categories[1].kind='gpt'],['category-size',v=>v.snapshot.categories.pop()],['category-total',v=>{v.snapshot.categories[0].started=1;v.snapshot.categories[0].active=1}],['bad-calendar',v=>v.snapshot.checkedAt='2026-02-31T00:00:00Z'],['bad-UTC',v=>v.snapshot.checkedAt='2026-10-13T08:01:00+08:00'],['start-future',v=>v.snapshot.startedAt='2026-10-13T00:02:00Z'],['export-before',v=>v.exportedAt='2026-10-13T00:01:00.123456788Z'],['process-pid',v=>v.snapshot.process.processId=0],['process-hash',v=>v.snapshot.process.programSHA256='SECRET'],['process-time',v=>v.snapshot.process.processStartedAt='2026-10-13T00:00:01.123456701Z']];
 for(const [name,change] of changes)await t.test(name,()=>{const v=report();change(v);assert.throws(()=>validateReplyReport(v));});
});

test('weighted averages and category maxima are checked while floating roundoff remains acceptable',()=>{
 const v=report(done);v.snapshot.averageFirstReplyMilliseconds=101;assert.throws(()=>validateReplyReport(v));v.snapshot.averageFirstReplyMilliseconds=100;v.snapshot.maxLastReplyMilliseconds=251;assert.throws(()=>validateReplyReport(v));
 const mixed=report();for(let i=0;i<3;i++)mixed.snapshot.categories[i]={kind:['gpt','chat','agent'][i],...done,started:1,completed:1,withReply:1,replyAttempts:1,replyAccepted:1,averageFirstReplyMilliseconds:100+i,maxFirstReplyMilliseconds:150+i,averageLastReplyMilliseconds:200+i,maxLastReplyMilliseconds:250+i};Object.assign(mixed.snapshot,{...done,averageFirstReplyMilliseconds:101+1e-12,maxFirstReplyMilliseconds:152,averageLastReplyMilliseconds:201,maxLastReplyMilliseconds:252});assert.equal(validateReplyReport(mixed).snapshot.categories.length,3);
});

test('changing collector identity precision or count maximum rejects pairing without fabricating deltas',async t=>{
 const changes=[['collector',f=>f.after.snapshot.startedAt='2026-10-13T00:00:01.123456701Z'],['pid',f=>f.after.snapshot.process.processId++],['created',f=>f.after.snapshot.process.processStartedAt='2026-10-13T00:00:00.000000101Z'],['hash',f=>f.after.snapshot.process.programSHA256='B'.repeat(64)],['identity-lost',f=>delete f.after.snapshot.process],['reverse-time',f=>f.after.snapshot.checkedAt='2026-10-13T00:01:00.123456788Z'],['same-time-changed',f=>f.after.snapshot.checkedAt=f.before.snapshot.checkedAt],['counter-rollback',f=>{f.before=report(done);f.after=report(zero,'2026-10-13T00:10:00Z')}],['max-rollback',f=>{f.before=report(done);f.after=report({...done,started:4,completed:4,withReply:4,replyAttempts:4,replyAccepted:4,maxFirstReplyMilliseconds:149},'2026-10-13T00:10:00Z')}]];
 for(const [name,change] of changes)await t.test(name,async()=>{const f=fixture();change(f);const r=await check(f);assert.equal(r.snapshotsValid,true);assert.equal(r.replyWindowValid,false);assert.equal(r.evidenceValid,false);assert.equal(r.window,undefined);assert.equal(replyEvidenceExitCode(r),1);});
});

test('latency cannot change without an ended reply while maxima may remain equal for faster new tasks',()=>{
 const a=validateReplyReport(report(done)).snapshot,b=clone(a);b.checkedAt='2026-10-13T00:02:00.123456789Z';b.averageFirstReplyMilliseconds=99;b.categories[0].averageFirstReplyMilliseconds=99;assert.throws(()=>replyDelta(a,b));Object.assign(b,{started:4,active:1});Object.assign(b.categories[0],{started:4,active:1});b.averageFirstReplyMilliseconds=a.averageFirstReplyMilliseconds;b.categories[0].averageFirstReplyMilliseconds=a.averageFirstReplyMilliseconds;assert.equal(replyDelta(a,b).started,1);
});

test('cumulative averages cannot lose prior latency or add more than the new maximum allows',async t=>{
 const changes=[['first-lower','averageFirstReplyMilliseconds',1],['last-lower','averageLastReplyMilliseconds',100],['first-upper','averageFirstReplyMilliseconds',149],['last-upper','averageLastReplyMilliseconds',249]];
 for(const [name,key,value] of changes)await t.test(name,()=>{
  const a=validateReplyReport(report(done)).snapshot,b=validateReplyReport(report({...done,started:4,completed:4,withReply:4,replyAttempts:4,replyAccepted:4},'2026-10-13T00:02:00Z')).snapshot;
  b[key]=value;b.categories[0][key]=value;assert.throws(()=>replyDelta(a,b));
 });
});

test('new receipts and completed outcomes must still correspond to active or newly ended tasks',()=>{
 const accepted=report({...zero,started:1,noReply:1,replyAccepted:1,replyAttempts:1});assert.throws(()=>validateReplyReport(accepted));const unconfirmed=report({...zero,started:1,noReply:1,replyUnconfirmed:1,replyAttempts:1});assert.throws(()=>validateReplyReport(unconfirmed));
 const a=validateReplyReport(report({...done,started:1,completed:0,failed:1,withReply:1,replyAttempts:1,replyAccepted:1})).snapshot;
 const b=clone(a);b.checkedAt='2026-10-13T00:02:00Z';b.started=2;b.completed=1;b.categories[0].started=2;b.categories[0].completed=1;assert.throws(()=>replyDelta(a,b));const c=clone(a);c.checkedAt=b.checkedAt;c.replyAttempts=2;c.replyAccepted=2;c.categories[0].replyAttempts=2;c.categories[0].replyAccepted=2;assert.throws(()=>replyDelta(a,c));const d=clone(a);d.checkedAt=b.checkedAt;d.replyAttempts=2;d.replyUnconfirmed=1;d.categories[0].replyAttempts=2;d.categories[0].replyUnconfirmed=1;assert.throws(()=>replyDelta(a,d));
});

test('new last reply timings cannot precede new first replies hidden by cumulative averages',()=>{
 const a=validateReplyReport(report(done)).snapshot,b=validateReplyReport(report({...done,started:4,completed:4,withReply:4,replyAttempts:4,replyAccepted:4,averageFirstReplyMilliseconds:110,averageLastReplyMilliseconds:150},'2026-10-13T00:02:00Z')).snapshot;assert.throws(()=>replyDelta(a,b));
});

test('safe integer boundary uses exact count conservation for tasks crossing the start',async()=>{
 const high={...zero,started:Number.MAX_SAFE_INTEGER,active:Number.MAX_SAFE_INTEGER},end={...done,started:Number.MAX_SAFE_INTEGER,active:Number.MAX_SAFE_INTEGER-2,completed:1,noReply:1,withReply:1,replyAttempts:1,replyAccepted:1};const r=await check({before:report(high),after:report(end,'2026-10-13T00:02:00Z')});assert.equal(replyEvidenceExitCode(r),0);assert.equal(r.window.started,0);assert.equal(r.window.completed,1);assert.equal(r.window.noReply,1);assert.equal(r.window.activeAtStart,Number.MAX_SAFE_INTEGER);
});

test('canonical ordering hash case and extra secrets never enter output',async()=>{
 const f=fixture();f.before.snapshot.categories.reverse();f.after.snapshot.process.programSHA256='a'.repeat(64);f.before.prompt='PRIVATE-TEXT';f.before.snapshot.account='PRIVATE-ACCOUNT';f.before.snapshot.process.path='PRIVATE-PATH';f.after.snapshot.categories[0].key='PRIVATE-KEY';const r=await check(f);assert.equal(replyEvidenceExitCode(r),0);assert.ok(!JSON.stringify(r).includes('PRIVATE'));assert.deepEqual(r.before.snapshot.categories.map(x=>x.kind),['gpt','chat','agent']);assert.equal(r.after.snapshot.process.programSHA256,'A'.repeat(64));
});

test('zero windows and missing process remain diagnostic while positive nanoseconds are preserved',async()=>{
 const f=fixture();f.after=clone(f.before);let r=await check(f);assert.equal(r.evidenceValid,true);assert.equal(replyEvidenceExitCode(r),2);assert.equal(r.window.durationNanoseconds,'0');delete f.before.snapshot.process;delete f.after.snapshot.process;f.after.snapshot.checkedAt='2026-10-13T00:01:00.123456790Z';r=await check(f);assert.equal(r.evidenceValid,true);assert.equal(r.window.durationNanoseconds,'1');assert.equal(r.window.durationSeconds,1e-9);assert.equal(r.snapshotProcessMatches,false);assert.equal(replyEvidenceExitCode(r),2);f.before.snapshot.process=report().snapshot.process;f.after.snapshot.process=clone(f.before.snapshot.process);r=await check(f);assert.equal(replyEvidenceExitCode(r),0);
});

test('full resource coverage matches exact identity and fractional memory without asserting business',async()=>{
 const input=await save(fixture()),resource=await saveResource(resourceFixture()),r=await checkReplyEvidence({...input,resource});assert.equal(r.evidenceValid,true);assert.equal(r.resourceChecksPassed,true);assert.equal(r.sameProcessVerified,true);assert.equal(r.compatible,true);assert.equal(r.replyObservationPassed,true);assert.equal(replyEvidenceExitCode(r),0);assert.equal(r.resource.resources.privateGrowthMiB,0.0078125);assert.equal(r.businessAcceptancePassed,false);assert.equal(r.temporalCoverage,'inside');assert.deepEqual(r.processComparison,{processIdMatches:true,processStartedAtMatches:true,programSHA256Matches:true});
});

test('resource timing cadence thresholds log clock and identity failures remain separate from valid snapshots',async t=>{
 const changes=[['short',f=>f.meta.requestedDurationSeconds=10],['stopped',f=>{f.meta.completed=false;f.meta.stopReason='stop-request'}],['sparse',f=>f.samples.splice(1,10)],['threshold',f=>f.meta.thresholds.privateGrowthMiB=0.001],['errors',f=>{f.meta.logCounts={...f.meta.logCounts,appendedLines:1,errorLines:1,bytesRead:50};f.meta.noNewErrorLines=false}],['coverage',f=>{f.meta.logCounts.unreadBytes=1;f.meta.logCounts.coverageComplete=false}],['clock',f=>{for(let i=1;i<f.samples.length;i++)f.samples[i].at=new Date(Date.parse(f.samples[i].at)+1500).toISOString();f.meta.completedAt='2026-10-13T00:15:04Z'}],['unstable',f=>f.meta.identityStable=false],['program',f=>f.meta.programUnchanged=false],['wrong-pid',f=>f.meta.processId++],['wrong-created',f=>f.meta.processStartedAt='2026-10-13T00:00:00.000000101Z'],['wrong-hash',f=>f.meta.programSHA256='B'.repeat(64)]];
 for(const [name,change] of changes)await t.test(name,async()=>{const f=resourceFixture();change(f);const resource=await saveResource(f),input=await save(fixture()),r=await checkReplyEvidence({...input,resource});assert.equal(r.snapshotsValid,true);assert.equal(r.replyWindowValid,true);assert.equal(r.evidenceValid,true);assert.equal(r.compatible,false);assert.equal(replyEvidenceExitCode(r),2);});
});

test('nonoverlapping partial zero and edge boundaries use nanosecond precision',async t=>{
 const resource=await saveResource(resourceFixture());const cases=[['none','2026-10-13T00:16:00Z','2026-10-13T00:17:00Z','none'],['partial','2026-10-13T00:00:01.999999999Z','2026-10-13T00:10:00Z','partial'],['zero','2026-10-13T00:01:00Z','2026-10-13T00:01:00Z','zero-window'],['edge','2026-10-13T00:00:02Z','2026-10-13T00:15:02Z','inside']];
 for(const [name,start,end,coverage] of cases)await t.test(name,async()=>{const f=fixture();f.before.snapshot.checkedAt=start;f.after.snapshot.checkedAt=end;if(name==='zero')f.after=clone(f.before);const r=await checkReplyEvidence({...await save(f),resource});assert.equal(r.evidenceValid,true);assert.equal(r.temporalCoverage,coverage);assert.equal(replyEvidenceExitCode(r),coverage==='inside'?0:2);});
});

test('corrupt and oversized snapshots or resource summaries return safe failures',async()=>{
 const input=await save(fixture());await writeFile(input.after,'PRIVATE-BAD-JSON');let r=await checkReplyEvidence(input);assert.equal(replyEvidenceExitCode(r),1);assert.ok(!JSON.stringify(r).includes('PRIVATE'));await writeFile(input.after,' '.repeat(1024*1024+1));r=await checkReplyEvidence(input);assert.equal(replyEvidenceExitCode(r),1);assert.ok(r.issues.includes('after:reply-json-size-limit'));const resource=await saveResource(resourceFixture()),meta=JSON.parse(await readFile(path.join(resource,'result.json'),'utf8'));meta.resources.privateGrowthMiB=0;await writeFile(path.join(resource,'result.json'),JSON.stringify(meta));r=await checkReplyEvidence({...await save(fixture()),resource});assert.equal(r.replyWindowValid,true);assert.equal(r.evidenceValid,false);assert.equal(replyEvidenceExitCode(r),1);assert.ok(r.issues.includes('resource:recorded-summary-mismatch'));
});

test('CLI outputs preserve inputs refuse existing files and directory aliases and omit secret errors',async()=>{
 const input=await save(fixture()),tool=fileURLToPath(new URL('check-reply-evidence.mjs',import.meta.url)),output=path.join(path.dirname(input.before),'checked.json'),run=args=>spawnSync(process.execPath,[tool,...args],{encoding:'utf8'}),args=['--before',input.before,'--after',input.after,'--output',output];const original=await readFile(input.before),first=run(args);assert.equal(first.status,0,first.stderr);assert.equal(JSON.parse(first.stdout).replyObservationPassed,true);assert.equal(run(args).status,1);assert.equal(run(['--before',input.before,'--after',input.after,'--output',input.before]).status,1);assert.deepEqual(await readFile(input.before),original);const alias=path.join(root,'alias-'+sequence);await symlink(path.dirname(input.before),alias,process.platform==='win32'?'junction':'dir');const link=path.join(alias,'before.json');assert.equal(run(['--before',input.before,'--after',input.after,'--output',link]).status,1);assert.deepEqual(await readFile(input.before),original);
 for(const args of [[],['--before','PRIVATE'],['--after','PRIVATE'],['--before',input.before,'--before',input.before,'--after',input.after],['--unknown','PRIVATE']]){const r=run(args);assert.equal(r.status,1);assert.ok(!r.stderr.includes('PRIVATE'))}
 const short=fixture();short.after=clone(short.before);const same=await save(short);assert.equal(run(['--before',same.before,'--after',same.after]).status,2);const resource=await saveResource(resourceFixture());assert.equal(run(['--before',input.before,'--after',input.after,'--resource',resource]).status,0);const bom=await save(fixture());await writeFile(bom.before,'\uFEFF'+await readFile(bom.before,'utf8'));assert.equal(run(['--before',bom.before,'--after',bom.after]).status,0);
});
