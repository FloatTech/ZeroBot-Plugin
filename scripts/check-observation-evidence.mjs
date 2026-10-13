// Offline evidence validation. No network, credentials, service control or source rewriting.
import {createReadStream} from 'node:fs';
import {open} from 'node:fs/promises';
import {createHash} from 'node:crypto';
import path from 'node:path';
import {pathToFileURL} from 'node:url';

const kinds=['gpt','chat','agent','vision'];
const counts=['started','active','completed','failed','incomplete'];
const ended=['started','completed','failed','incomplete'];
const deltaKeys=[...ended,'activeAtStart','activeAtEnd'];
const MAX_JSON=1024*1024, MAX_SAMPLES=64*1024*1024;
class InvalidEvidence extends Error { constructor(code){super(code);this.code=code;} }
function requireEvidence(condition,code){if(!condition)throw new InvalidEvidence(code);}
const integer=value=>Number.isSafeInteger(value)&&value>=0;
const nonnegative=value=>typeof value==='number'&&Number.isFinite(value)&&value>=0;
const sum=(row,keys)=>keys.reduce((n,key)=>n+BigInt(row[key]),0n);
const close=(a,b)=>typeof a==='number'&&typeof b==='number'&&Number.isFinite(a)&&Number.isFinite(b)&&Math.abs(a-b)<=Math.max(1e-9,Math.abs(b)*1e-10);

export function stamp(value,utcOnly=false){
  requireEvidence(typeof value==='string','invalid-time');
  const m=value.match(/^(\d{4})-(\d{2})-(\d{2})T(\d{2}):(\d{2}):(\d{2})(?:\.(\d{1,9}))?(Z|([+-])(\d{2}):(\d{2}))$/);
  requireEvidence(m&&(!utcOnly||m[8]==='Z'),'invalid-time');
  const [year,month,day,hour,minute,second]=m.slice(1,7).map(Number);
  requireEvidence(year>=2000&&month>=1&&month<=12&&day>=1&&hour<=23&&minute<=59&&second<=59,'invalid-time');
  const base=Date.UTC(year,month-1,day,hour,minute,second),d=new Date(base);
  requireEvidence(d.getUTCFullYear()===year&&d.getUTCMonth()===month-1&&d.getUTCDate()===day,'invalid-time');
  const offsetHour=Number(m[10]||0),offsetMinute=Number(m[11]||0);
  requireEvidence(offsetHour<=14&&offsetMinute<=59&&(offsetHour!==14||offsetMinute===0),'invalid-time');
  const offset=(offsetHour*60+offsetMinute)*60000*(m[9]==='-'?-1:1);
  const fraction=(m[7]||'').padEnd(9,'0'),ns=BigInt(base-offset)*1000000n+BigInt(fraction);
  return {ns,ms:base-offset+Number(fraction.slice(0,3)),text:new Date(base-offset).toISOString().slice(0,19)+'.'+fraction+'Z'};
}

function row(value){
  requireEvidence(value&&counts.every(key=>integer(value[key])),'invalid-model-counts');
  requireEvidence(BigInt(value.started)===sum(value,['active','completed','failed','incomplete']),'model-count-conservation');
  requireEvidence(nonnegative(value.averageMilliseconds)&&nonnegative(value.maxMilliseconds)&&value.averageMilliseconds<=value.maxMilliseconds,'invalid-model-latency');
  requireEvidence(sum(value,['completed','failed','incomplete'])!==0n||(value.averageMilliseconds===0&&value.maxMilliseconds===0),'invalid-model-latency');
  return Object.fromEntries([...counts,'averageMilliseconds','maxMilliseconds'].map(key=>[key,value[key]]));
}
function categories(value,copy){
  requireEvidence(Array.isArray(value)&&value.length===4,'invalid-model-categories');
  return kinds.map(kind=>{
    const matches=value.filter(item=>item?.kind===kind);
    requireEvidence(matches.length===1,'invalid-model-categories');
    return {kind,...copy(matches[0])};
  });
}
function snapshot(value,withProcess=false){
  const totals=row(value),start=stamp(value.startedAt,true),checked=stamp(value.checkedAt,true);
  requireEvidence(checked.ns>=start.ns,'model-time-order');
  const rows=categories(value.categories,row);
  requireEvidence(counts.every(key=>BigInt(totals[key])===rows.reduce((n,r)=>n+BigInt(r[key]),0n)),'model-category-totals');
  let process;
  if(withProcess){
    const p=value.process;
    requireEvidence(p&&integer(p.processId)&&p.processId>0&&typeof p.programSHA256==='string'&&/^[a-fA-F0-9]{64}$/.test(p.programSHA256),'invalid-model-process');
    const processStart=stamp(p.processStartedAt,true);
    requireEvidence(processStart.ns<=start.ns,'model-process-time-order');
    process={processId:p.processId,processStartedAt:processStart.text,programSHA256:p.programSHA256.toUpperCase()};
  }
  return {startedAt:start.text,checkedAt:checked.text,...totals,categories:rows,...(process?{process}:{})};
}
function delta(before,after){
  requireEvidence(before.startedAt===after.startedAt,'model-collector-changed');
  requireEvidence(JSON.stringify(before.process||null)===JSON.stringify(after.process||null),'model-process-changed');
  requireEvidence(stamp(after.checkedAt).ns>=stamp(before.checkedAt).ns,'model-time-order');
  requireEvidence(before.checkedAt!==after.checkedAt||JSON.stringify(before)===JSON.stringify(after),'model-same-time-changed');
  const compare=(a,b)=>{
    requireEvidence(ended.every(key=>b[key]>=a[key])&&b.maxMilliseconds>=a.maxMilliseconds,'model-counter-rollback');
    const value=Object.fromEntries(ended.map(key=>[key,b[key]-a[key]]));
    requireEvidence(BigInt(value.started)+BigInt(a.active)===sum(value,['completed','failed','incomplete'])+BigInt(b.active),'model-window-conservation');
    return {...value,activeAtStart:a.active,activeAtEnd:b.active};
  };
  return {durationSeconds:(stamp(after.checkedAt).ms-stamp(before.checkedAt).ms)/1000,...compare(before,after),categories:kinds.map((kind,i)=>({kind,...compare(before.categories[i],after.categories[i])}))};
}
export function validateModel(value){
  requireEvidence(value?.format==='zerobot-model-observation'&&[1,2].includes(value.version),'unsupported-model-format');
  requireEvidence(value.scope?.kind==='model-http-exchanges'&&value.scope.latencyScope==='process-cumulative'&&value.scope.businessAcceptancePassed===false,'invalid-model-scope');
  const before=snapshot(value.before,value.version===2),after=snapshot(value.after,value.version===2),window=delta(before,after),exported=stamp(value.exportedAt,true);
  requireEvidence(exported.ns>=stamp(after.checkedAt).ns,'model-export-time-order');
  const copyDelta=v=>{requireEvidence(v&&deltaKeys.every(key=>integer(v[key])),'invalid-model-window');return Object.fromEntries(deltaKeys.map(key=>[key,v[key]]));};
  const actual={durationSeconds:value.window?.durationSeconds,...copyDelta(value.window),categories:categories(value.window?.categories,copyDelta)};
  requireEvidence(JSON.stringify(actual)===JSON.stringify(window),'model-window-mismatch');
  requireEvidence(!['averageMilliseconds','maxMilliseconds'].some(key=>Object.hasOwn(value.window,key)||value.window.categories.some(r=>Object.hasOwn(r,key))),'model-window-latency-scope');
  return {version:value.version,exportedAt:exported.text,before,after,window};
}

async function readJSON(filename){
  const chunks=[],hash=createHash('sha256');let size=0;
  for await(const chunk of createReadStream(filename)){
    size+=chunk.length;requireEvidence(size<=MAX_JSON,'json-size-limit');chunks.push(chunk);hash.update(chunk);
  }
  return {value:JSON.parse(Buffer.concat(chunks).toString('utf8').replace(/^\uFEFF/,'')),sha256:hash.digest('hex')};
}
async function readSamples(filename){
  const samples=[],hash=createHash('sha256');let pending=Buffer.alloc(0),size=0;
  const line=buffer=>{
    requireEvidence(buffer.length<=4096,'sample-line-limit');
    const text=buffer.toString('utf8').replace(/\r$/,'');
    requireEvidence(text.trim().length>0,'empty-sample-line');
    requireEvidence(samples.length<100000,'sample-count-limit');
    const value=JSON.parse(text);
    requireEvidence(value&&nonnegative(value.atSeconds)&&nonnegative(value.cpuSeconds)&&['privateBytes','workingSetBytes','threadCount','handleCount'].every(key=>integer(value[key])),'invalid-sample-values');
    const time=stamp(value.at);
    samples.push({at:time.text,ns:time.ns,...Object.fromEntries(['atSeconds','cpuSeconds','privateBytes','workingSetBytes','threadCount','handleCount'].map(key=>[key,value[key]]))});
  };
  for await(const chunk of createReadStream(filename)){
    size+=chunk.length;requireEvidence(size<=MAX_SAMPLES,'sample-size-limit');hash.update(chunk);
    const buffer=Buffer.concat([pending,chunk]);let offset=0,index;
    while((index=buffer.indexOf(10,offset))!==-1){line(buffer.subarray(offset,index));offset=index+1;}
    pending=buffer.subarray(offset);requireEvidence(pending.length<=4096,'sample-line-limit');
  }
  if(pending.length)line(pending);
  requireEvidence(samples.length>=2,'sample-count-insufficient');
  return {samples,sha256:hash.digest('hex')};
}

export function validateResource(meta,samples){
  requireEvidence(meta&&integer(meta.processId)&&meta.processId>0&&typeof meta.programSHA256==='string'&&/^[A-Fa-f0-9]{64}$/.test(meta.programSHA256),'invalid-resource-identity');
  const start=stamp(meta.startedAt),end=stamp(meta.completedAt),processStart=stamp(meta.processStartedAt);
  requireEvidence(processStart.ns<=start.ns&&start.ns<=samples[0].ns&&samples.at(-1).ns<=end.ns,'resource-time-order');
  requireEvidence(integer(meta.requestedDurationSeconds)&&meta.requestedDurationSeconds>=10&&meta.requestedDurationSeconds<=86400&&integer(meta.intervalSeconds)&&meta.intervalSeconds>=1&&meta.intervalSeconds<=30&&integer(meta.logicalCPUs)&&meta.logicalCPUs>0,'invalid-resource-parameters');
  requireEvidence(nonnegative(meta.thresholds?.cpuAverageCapacityPercent)&&meta.thresholds.cpuAverageCapacityPercent>0&&nonnegative(meta.thresholds.privateGrowthMiB)&&meta.thresholds.privateGrowthMiB>0,'invalid-resource-thresholds');
  requireEvidence(['completed','passed','identityStable','programUnchanged','noNewErrorLines'].every(key=>typeof meta[key]==='boolean'),'invalid-resource-flags');
  requireEvidence(['duration-complete','stop-request','process-identity-changed','observation-error'].includes(meta.stopReason)&&meta.completed===(meta.stopReason==='duration-complete'),'invalid-resource-stop-reason');
  requireEvidence(meta.realQQMessagesSent===false&&meta.modelCalls===0&&meta.businessLoadMeasured===false&&meta.businessAcceptancePassed===false&&meta.messageContentsSaved===false&&meta.servicesStartedOrStopped===false,'invalid-resource-scope');
  const log=meta.logCounts,logKeys=['appendedLines','warningLines','errorLines','fatalLines','panicLines','otherLines','bytesRead','unreadBytes','partialLines'];
  requireEvidence(log&&logKeys.every(key=>integer(log[key]))&&typeof log.logTruncated==='boolean'&&typeof log.coverageComplete==='boolean','invalid-log-summary');
  requireEvidence(BigInt(log.appendedLines)===sum(log,['warningLines','errorLines','fatalLines','panicLines','otherLines'])&&log.appendedLines<=log.bytesRead,'log-count-conservation');
  requireEvidence(log.coverageComplete===(!log.logTruncated&&log.unreadBytes===0&&log.partialLines===0)&&meta.noNewErrorLines===(log.errorLines===0&&log.fatalLines===0&&log.panicLines===0),'log-flags-mismatch');
  const first=samples[0],last=samples.at(-1);let gapMax=0,cpuPeak=0,clockDrift=0;
  let privateMax=first.privateBytes,workingMax=first.workingSetBytes,threadMax=first.threadCount,handleMax=first.handleCount;
  for(let i=0;i<samples.length;i++){
    const r=samples[i];
    clockDrift=Math.max(clockDrift,Math.abs(Number(r.ns-first.ns)/1e9-(r.atSeconds-first.atSeconds)));
    privateMax=Math.max(privateMax,r.privateBytes);workingMax=Math.max(workingMax,r.workingSetBytes);threadMax=Math.max(threadMax,r.threadCount);handleMax=Math.max(handleMax,r.handleCount);
    if(i){const prev=samples[i-1],gap=r.atSeconds-prev.atSeconds,used=r.cpuSeconds-prev.cpuSeconds;
      requireEvidence(gap>0&&used>=0&&r.ns>prev.ns,'sample-counter-or-time-rollback');
      gapMax=Math.max(gapMax,gap);cpuPeak=Math.max(cpuPeak,used/gap/meta.logicalCPUs*100);
    }
  }
  const span=last.atSeconds-first.atSeconds,cpuAverage=(last.cpuSeconds-first.cpuSeconds)/span/meta.logicalCPUs*100,growth=(privateMax-first.privateBytes)/1048576;
  const resources={sampleCount:samples.length,sampleSpanSeconds:span,maxSampleGapSeconds:gapMax,cpuAveragePercent:cpuAverage,cpuPeakSamplePercent:cpuPeak,privateFirstBytes:first.privateBytes,privateMaxBytes:privateMax,privateLastBytes:last.privateBytes,privateGrowthMiB:growth,workingSetMaxBytes:workingMax,threadMax,handleMax,metricsValid:true,completeWindow:span>=meta.requestedDurationSeconds,cadencePassed:gapMax<=2*meta.intervalSeconds+0.5&&samples.length>=Math.ceil(meta.requestedDurationSeconds/meta.intervalSeconds*0.8)+1,resourceThresholdsPassed:cpuAverage<=meta.thresholds.cpuAverageCapacityPercent&&growth<=meta.thresholds.privateGrowthMiB,fullTiming:meta.requestedDurationSeconds>=900};
  const mismatches=Object.keys(resources).filter(key=>typeof resources[key]==='boolean'?meta.resources?.[key]!==resources[key]:integer(resources[key])&&['sampleCount','privateFirstBytes','privateMaxBytes','privateLastBytes','workingSetMaxBytes','threadMax','handleMax'].includes(key)?meta.resources?.[key]!==resources[key]:!close(meta.resources?.[key],resources[key]));
  const passed=meta.completed&&meta.identityStable&&meta.programUnchanged&&resources.fullTiming&&resources.completeWindow&&resources.cadencePassed&&resources.resourceThresholdsPassed&&meta.noNewErrorLines&&log.coverageComplete&&meta.stopReason!=='observation-error';
  if(meta.passed!==passed)mismatches.push('passed');
  return {processId:meta.processId,processStartedAt:processStart.text,programSHA256:meta.programSHA256.toUpperCase(),firstSampleAt:first.at,lastSampleAt:last.at,requestedDurationSeconds:meta.requestedDurationSeconds,intervalSeconds:meta.intervalSeconds,logicalCPUs:meta.logicalCPUs,thresholds:{cpuAverageCapacityPercent:meta.thresholds.cpuAverageCapacityPercent,privateGrowthMiB:meta.thresholds.privateGrowthMiB},stopReason:meta.stopReason,resources,completed:meta.completed,identityStable:meta.identityStable,programUnchanged:meta.programUnchanged,logCounts:Object.fromEntries([...logKeys,'logTruncated','coverageComplete'].map(key=>[key,log[key]])),recordedSummaryMatches:mismatches.length===0,summaryMismatches:mismatches,clockConsistent:clockDrift<=1,maxClockDriftSeconds:clockDrift,resourceChecksPassed:passed};
}

async function loadResourceEvidence(output,resource){
  try {
    const input=await readJSON(path.join(resource,'result.json'));output.sources.resourceSHA256=input.sha256;
    const raw=await readSamples(path.join(resource,'samples.jsonl'));output.sources.samplesSHA256=raw.sha256;
    const result=validateResource(input.value,raw.samples);output.resource=result;output.resourceChecksPassed=result.resourceChecksPassed;
    if(!result.recordedSummaryMatches)output.issues.push('resource:recorded-summary-mismatch');
    return result;
  }catch(error){output.issues.push('resource:'+(error instanceof InvalidEvidence?error.code:'file-or-json-unreadable'));return null;}
}

export async function checkResourceEvidence({resource}){
  const output={format:'zerobot-observation-evidence-check',version:1,mode:'resource-only',checkedAt:new Date().toISOString(),evidenceValid:false,resourceChecksPassed:false,resourceObservationPassed:false,temporalCoverage:'unavailable',compatible:false,sameProcessVerified:false,businessAcceptancePassed:false,issues:[],sources:{}};
  const result=await loadResourceEvidence(output,resource);
  output.evidenceValid=!!result&&result.recordedSummaryMatches;
  output.resourceObservationPassed=output.evidenceValid&&output.resourceChecksPassed&&result.clockConsistent;
  if(result&&!result.clockConsistent)output.issues.push('resource:sample-clock-drift');
  if(result&&!output.resourceChecksPassed)output.issues.push('resource:checks-not-passed');
  output.issues.push('association:model-report-not-provided','acceptance:real-business-not-verified');
  return output;
}

export async function checkEvidence({resource,model}){
  const output={format:'zerobot-observation-evidence-check',version:1,checkedAt:new Date().toISOString(),evidenceValid:false,resourceChecksPassed:false,temporalCoverage:'unavailable',compatible:false,sameProcessVerified:false,businessAcceptancePassed:false,issues:[],sources:{}};
  const failure=(scope,error)=>output.issues.push(scope+':'+(error instanceof InvalidEvidence?error.code:'file-or-json-unreadable'));
  const [resourceInput,modelInput]=await Promise.allSettled([loadResourceEvidence(output,resource),readJSON(model)]);
  const resourceResult=resourceInput.status==='fulfilled'?resourceInput.value:null;let modelResult;
  if(resourceInput.status==='rejected')failure('resource',resourceInput.reason);
  if(modelInput.status==='fulfilled'){
    output.sources.modelSHA256=modelInput.value.sha256;
    try{modelResult=validateModel(modelInput.value.value);output.model=modelResult;}catch(error){failure('model',error);}
  }else failure('model',modelInput.reason);
  output.evidenceValid=!!resourceResult&&!!modelResult&&resourceResult.recordedSummaryMatches;
  if(resourceResult&&modelResult){
    const a=stamp(modelResult.before.checkedAt).ns,b=stamp(modelResult.after.checkedAt).ns,first=stamp(resourceResult.firstSampleAt).ns,last=stamp(resourceResult.lastSampleAt).ns;
    output.temporalCoverage=b<first||a>last?'none':modelResult.window.durationSeconds===0?'zero-window':a>=first&&b<=last?'inside':'partial';
    const collectorPossible=stamp(modelResult.before.startedAt).ns>=stamp(resourceResult.processStartedAt).ns;
    if(output.temporalCoverage!=='inside')output.issues.push('association:window-'+output.temporalCoverage);
    if(!collectorPossible)output.issues.push('association:collector-before-process');
    if(!resourceResult.clockConsistent)output.issues.push('association:sample-clock-drift');
    if(!resourceResult.resourceChecksPassed)output.issues.push('resource:checks-not-passed');
    output.compatible=output.evidenceValid&&output.resourceChecksPassed&&output.temporalCoverage==='inside'&&collectorPossible&&resourceResult.clockConsistent;
    if(modelResult.version===2){
      const p=modelResult.before.process;
      output.processComparison={processIdMatches:p.processId===resourceResult.processId,processStartedAtMatches:p.processStartedAt===resourceResult.processStartedAt,programSHA256Matches:p.programSHA256===resourceResult.programSHA256};
      const matches=Object.values(output.processComparison).every(Boolean);
      output.sameProcessVerified=output.evidenceValid&&matches&&resourceResult.identityStable&&resourceResult.programUnchanged;
      if(!matches)output.issues.push('association:process-identity-mismatch');
      if(!resourceResult.identityStable||!resourceResult.programUnchanged)output.issues.push('association:resource-process-not-stable');
      output.compatible=output.compatible&&output.sameProcessVerified;
    }
  }
  if(modelResult?.version!==2)output.issues.push('association:model-report-has-no-process-identity');
  output.issues.push('acceptance:real-business-not-verified');
  return output;
}

export function evidenceExitCode(result){return !result.evidenceValid?1:(result.mode==='resource-only'?result.resourceObservationPassed:result.compatible)?0:2;}
export async function writeEvidence(filename,result){
  // Exclusive creation protects existing output, inputs and symlink targets alike.
  const handle=await open(filename,'wx',0o600);
  try{await handle.writeFile(JSON.stringify(result,null,2)+'\n','utf8');}finally{await handle.close();}
}
export async function main(args=process.argv.slice(2)){
  try{
    const options={};
    for(let i=0;i<args.length;i+=2){const key=args[i];requireEvidence(['--resource','--model','--output'].includes(key)&&args[i+1]&&!args[i+1].startsWith('--')&&!Object.hasOwn(options,key.slice(2)),'invalid-arguments');options[key.slice(2)]=args[i+1];}
    requireEvidence(options.resource,'invalid-arguments');
    const result=await (options.model?checkEvidence(options):checkResourceEvidence(options));if(options.output)await writeEvidence(options.output,result);
    process.stdout.write(JSON.stringify(result,null,2)+'\n');return evidenceExitCode(result);
  }catch{process.stderr.write('观察核对失败：请检查参数、文件读取权限，以及输出文件是否尚不存在。\n');return 1;}
}
if(process.argv[1]&&import.meta.url===pathToFileURL(path.resolve(process.argv[1])).href)process.exitCode=await main();
