// Offline saved reply snapshots only. Never access credentials, APIs or service controls.
import {createReadStream} from 'node:fs';
import {createHash} from 'node:crypto';
import path from 'node:path';
import {pathToFileURL} from 'node:url';
import {stamp,checkResourceEvidence,writeEvidence} from './check-observation-evidence.mjs';

const kinds=['gpt','chat','agent'];
const counts=['started','active','completed','failed','noReply','withReply','missingIngress','replyAttempts','replyAccepted','replyUnconfirmed'];
const cumulative=counts.filter(k=>k!=='active');
const times=['averageFirstReplyMilliseconds','maxFirstReplyMilliseconds','averageLastReplyMilliseconds','maxLastReplyMilliseconds'];
const boundary='bot-matcher-entry-to-onebot-send-receipt';
class InvalidReply extends Error {constructor(code){super(code);this.code=code;}}
const requireReply=(condition,code)=>{if(!condition)throw new InvalidReply(code);};
const integer=v=>Number.isSafeInteger(v)&&v>=0;
const sum=(v,keys)=>keys.reduce((n,k)=>n+BigInt(v[k]),0n);
const close=(a,b)=>Math.abs(a-b)<=Math.max(1e-9,Math.abs(b)*1e-10);
function utc(value){try{return stamp(value,true);}catch{throw new InvalidReply('invalid-reply-time');}}
function row(value){
 requireReply(value&&counts.every(k=>integer(value[k])),'invalid-reply-counts');
 requireReply(BigInt(value.started)===sum(value,['active','completed','failed','noReply'])&&BigInt(value.replyAttempts)===sum(value,['replyAccepted','replyUnconfirmed']),'reply-count-conservation');
 requireReply(value.missingIngress<=value.started&&value.completed<=value.withReply&&BigInt(value.withReply)<=sum(value,['completed','failed'])&&value.withReply<=value.replyAccepted,'invalid-reply-denominator');
 requireReply((value.replyAccepted===0||value.withReply>0||value.active>0)&&(value.replyUnconfirmed===0||value.failed>0||value.active>0),'reply-receipts-without-task');
 requireReply(times.every(k=>typeof value[k]==='number'&&Number.isFinite(value[k])&&value[k]>=0)&&value.averageFirstReplyMilliseconds<=value.maxFirstReplyMilliseconds&&value.averageLastReplyMilliseconds<=value.maxLastReplyMilliseconds&&value.averageLastReplyMilliseconds>=value.averageFirstReplyMilliseconds&&value.maxLastReplyMilliseconds>=value.maxFirstReplyMilliseconds,'invalid-reply-latency');
 requireReply(value.withReply!==0||times.every(k=>value[k]===0),'empty-reply-latency');
 return Object.fromEntries([...counts,...times].map(k=>[k,value[k]]));
}
export function validateReplyReport(value){
 requireReply(value?.format==='zerobot-reply-statistics'&&value.version===1,'unsupported-reply-format');
 requireReply(value.scope?.kind==='ai-reply-tasks'&&value.scope.latencyScope==='process-cumulative'&&value.scope.timingBoundary===boundary&&value.scope.businessAcceptancePassed===false,'invalid-reply-scope');
 requireReply(!Object.hasOwn(value,'window'),'unexpected-reply-window');
 const source=value.snapshot,total=row(source),started=utc(source.startedAt),checked=utc(source.checkedAt),exported=utc(value.exportedAt);
 requireReply(source.scope===boundary&&started.ns<=checked.ns&&checked.ns<=exported.ns,'reply-time-or-boundary');
 requireReply(Array.isArray(source.categories)&&source.categories.length===3,'invalid-reply-categories');
 const categories=kinds.map(kind=>{const matches=source.categories.filter(x=>x?.kind===kind);requireReply(matches.length===1,'invalid-reply-categories');return {kind,...row(matches[0])};});
 requireReply(counts.every(k=>BigInt(total[k])===categories.reduce((n,r)=>n+BigInt(r[k]),0n)),'reply-category-totals');
 for(const k of ['maxFirstReplyMilliseconds','maxLastReplyMilliseconds'])requireReply(total[k]===Math.max(...categories.map(r=>r[k])),'reply-category-latency');
 for(const k of ['averageFirstReplyMilliseconds','averageLastReplyMilliseconds']){
  const average=total.withReply?categories.reduce((n,r)=>n+r[k]*(r.withReply/total.withReply),0):0;
  requireReply(Number.isFinite(average)&&close(total[k],average),'reply-category-latency');
 }
 let process;
 if(source.process!=null){
  const p=source.process;requireReply(integer(p.processId)&&p.processId>0&&typeof p.programSHA256==='string'&&/^[a-fA-F0-9]{64}$/.test(p.programSHA256),'invalid-reply-process');
  const created=utc(p.processStartedAt);requireReply(created.ns<=started.ns,'reply-process-time-order');
  process={processId:p.processId,processStartedAt:created.text,programSHA256:p.programSHA256.toUpperCase()};
 }
 return {exportedAt:exported.text,snapshot:{startedAt:started.text,checkedAt:checked.text,scope:boundary,...total,categories,...(process?{process}:{})}};
}
export function replyDelta(before,after){
 requireReply(before.startedAt===after.startedAt,'reply-collector-changed');
 requireReply(JSON.stringify(before.process||null)===JSON.stringify(after.process||null),'reply-process-changed');
 const duration=utc(after.checkedAt).ns-utc(before.checkedAt).ns;
 requireReply(duration>=0n,'reply-reading-time-reversed');
 requireReply(duration!==0n||JSON.stringify(before)===JSON.stringify(after),'reply-same-time-changed');
 const compare=(a,b)=>{
  requireReply(cumulative.every(k=>b[k]>=a[k])&&b.maxFirstReplyMilliseconds>=a.maxFirstReplyMilliseconds&&b.maxLastReplyMilliseconds>=a.maxLastReplyMilliseconds,'reply-counter-or-maximum-rollback');
  const d=Object.fromEntries(cumulative.map(k=>[k,b[k]-a[k]]));
  requireReply(BigInt(d.started)+BigInt(a.active)===sum(d,['completed','failed','noReply'])+BigInt(b.active),'reply-window-conservation');
  requireReply(d.missingIngress<=d.started&&d.completed<=d.withReply&&BigInt(d.withReply)<=sum(d,['completed','failed'])&&BigInt(d.replyAttempts)===sum(d,['replyAccepted','replyUnconfirmed']),'reply-window-denominator');
  requireReply((d.replyAccepted===0||d.withReply>0||b.active>0)&&(d.replyUnconfirmed===0||d.failed>0||b.active>0),'reply-window-receipts-without-task');
  requireReply(d.withReply!==0||times.every(k=>a[k]===b[k]),'reply-latency-without-ended-reply');
  if(d.withReply>0){
   for(const [average,maximum] of [['averageFirstReplyMilliseconds','maxFirstReplyMilliseconds'],['averageLastReplyMilliseconds','maxLastReplyMilliseconds']]){
    const lower=a[average]*(a.withReply/b.withReply),upper=lower+b[maximum]*(d.withReply/b.withReply);
    requireReply((b[average]>=lower||close(b[average],lower))&&(b[average]<=upper||close(b[average],upper)),'reply-cumulative-latency-inconsistent');
   }
   const prior=a.withReply/b.withReply,first=b.averageFirstReplyMilliseconds-a.averageFirstReplyMilliseconds*prior,last=b.averageLastReplyMilliseconds-a.averageLastReplyMilliseconds*prior;
   const tolerance=Math.max(1e-9,...[a.averageFirstReplyMilliseconds,a.averageLastReplyMilliseconds,b.averageFirstReplyMilliseconds,b.averageLastReplyMilliseconds].map(x=>Math.abs(x)*1e-10));
   requireReply(last>=first||first-last<=tolerance,'reply-incremental-first-last-order');
  }
  return {...d,activeAtStart:a.active,activeAtEnd:b.active};
 };
 return {durationSeconds:Number(duration)/1e9,durationNanoseconds:duration.toString(),...compare(before,after),categories:kinds.map((kind,i)=>({kind,...compare(before.categories[i],after.categories[i])}))};
}
async function readReply(filename){
 const chunks=[],hash=createHash('sha256');let size=0;
 for await(const chunk of createReadStream(filename)){size+=chunk.length;requireReply(size<=1024*1024,'reply-json-size-limit');chunks.push(chunk);hash.update(chunk);}
 return {value:JSON.parse(Buffer.concat(chunks).toString('utf8').replace(/^\uFEFF/,'')),sha256:hash.digest('hex')};
}
export async function checkReplyEvidence({before,after,resource}){
 const output={format:'zerobot-reply-evidence-check',version:1,checkedAt:new Date().toISOString(),snapshotsValid:false,replyWindowValid:false,evidenceValid:false,replyObservationPassed:false,resourceProvided:resource!==undefined,resourceChecksPassed:false,temporalCoverage:'unavailable',snapshotProcessMatches:false,sameProcessVerified:false,compatible:false,businessAcceptancePassed:false,issues:[],sources:{}};
 const failure=(name,error)=>output.issues.push(name+':'+(error instanceof InvalidReply?error.code:'file-or-json-unreadable'));
 const inputs=await Promise.allSettled([readReply(before),readReply(after)]);const reports=[];
 for(let i=0;i<2;i++){
  const name=i===0?'before':'after',input=inputs[i];
  if(input.status==='rejected'){failure(name,input.reason);continue;}
  output.sources[name+'SHA256']=input.value.sha256;
  try{reports[i]=validateReplyReport(input.value.value);output[name]=reports[i];}catch(error){failure(name,error);}
 }
 output.snapshotsValid=!!reports[0]&&!!reports[1];
 if(output.snapshotsValid){
  try{
   output.window=replyDelta(reports[0].snapshot,reports[1].snapshot);output.replyWindowValid=true;
   output.snapshotProcessMatches=!!reports[0].snapshot.process;
   if(output.window.durationNanoseconds==='0')output.issues.push('reply:zero-window');
   if(!output.snapshotProcessMatches)output.issues.push('reply:process-identity-unavailable');
  }catch(error){failure('pair',error);}
 }
 output.evidenceValid=output.replyWindowValid;
 const usable=output.replyWindowValid&&output.snapshotProcessMatches&&output.window.durationNanoseconds!=='0';
 output.replyObservationPassed=usable;
 if(resource!==undefined){
  const checked=await checkResourceEvidence({resource});
  output.sources.resourceSHA256=checked.sources.resourceSHA256;output.sources.samplesSHA256=checked.sources.samplesSHA256;
  if(checked.resource)output.resource=checked.resource;
  output.resourceChecksPassed=checked.resourceChecksPassed;
  output.issues.push(...checked.issues.filter(x=>x.startsWith('resource:')));
  output.evidenceValid=output.replyWindowValid&&checked.evidenceValid;
  if(output.replyWindowValid&&checked.resource){
   const r=checked.resource,a=utc(reports[0].snapshot.checkedAt).ns,b=utc(reports[1].snapshot.checkedAt).ns,first=stamp(r.firstSampleAt).ns,last=stamp(r.lastSampleAt).ns;
   output.temporalCoverage=b<first||a>last?'none':a===b?'zero-window':a>=first&&b<=last?'inside':'partial';
   if(output.temporalCoverage!=='inside')output.issues.push('association:window-'+output.temporalCoverage);
   const p=reports[0].snapshot.process;
   if(p){
    output.processComparison={processIdMatches:p.processId===r.processId,processStartedAtMatches:p.processStartedAt===r.processStartedAt,programSHA256Matches:p.programSHA256===r.programSHA256};
    output.sameProcessVerified=output.evidenceValid&&Object.values(output.processComparison).every(Boolean)&&r.identityStable&&r.programUnchanged;
    if(!Object.values(output.processComparison).every(Boolean))output.issues.push('association:process-identity-mismatch');
   }
   if(!r.identityStable||!r.programUnchanged)output.issues.push('association:resource-process-not-stable');
   output.compatible=output.evidenceValid&&usable&&checked.resourceObservationPassed&&output.temporalCoverage==='inside'&&output.sameProcessVerified;
  }
  output.replyObservationPassed=output.compatible;
 }
 output.issues.push('acceptance:real-business-not-verified');return output;
}
export const replyEvidenceExitCode=r=>!r.evidenceValid?1:r.replyObservationPassed?0:2;
export async function main(args=process.argv.slice(2)){
 try{
  const options={};
  for(let i=0;i<args.length;i+=2){const key=args[i];requireReply(['--before','--after','--resource','--output'].includes(key)&&args[i+1]&&!args[i+1].startsWith('--')&&!Object.hasOwn(options,key.slice(2)),'invalid-arguments');options[key.slice(2)]=args[i+1];}
  requireReply(options.before&&options.after,'invalid-arguments');
  const result=await checkReplyEvidence(options);if(options.output)await writeEvidence(options.output,result);process.stdout.write(JSON.stringify(result,null,2)+'\n');return replyEvidenceExitCode(result);
 }catch{process.stderr.write('回复快照核对失败：请检查参数、资料读取权限，以及输出文件是否尚不存在。\n');return 1;}
}
if(process.argv[1]&&import.meta.url===pathToFileURL(path.resolve(process.argv[1])).href)process.exitCode=await main();
