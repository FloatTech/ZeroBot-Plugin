// Saved evidence only. Do not access credentials, networks or running services.
import path from 'node:path';
import {pathToFileURL} from 'node:url';
import {checkEvidence,stamp,writeEvidence} from './check-observation-evidence.mjs';
import {checkReplyEvidence} from './check-reply-evidence.mjs';

export function combineObservation(model,reply){
 const result={format:'zerobot-combined-observation-check',version:1,checkedAt:new Date().toISOString(),evidenceValid:false,resourceChecksPassed:false,sameProcessVerified:false,windowsOverlap:false,compatible:false,businessAcceptancePassed:false,issues:[],sources:{},model,reply};
 const sameResources=['resourceSHA256','samplesSHA256'].every(k=>typeof model.sources?.[k]==='string'&&model.sources[k]===reply.sources?.[k]);
 result.sources={resourceSHA256:model.sources?.resourceSHA256,samplesSHA256:model.sources?.samplesSHA256,modelSHA256:model.sources?.modelSHA256,beforeReplySHA256:reply.sources?.beforeSHA256,afterReplySHA256:reply.sources?.afterSHA256};
 result.evidenceValid=model.evidenceValid===true&&reply.evidenceValid===true&&sameResources;
 result.resourceChecksPassed=model.resourceChecksPassed===true&&reply.resourceChecksPassed===true;
 result.sameProcessVerified=result.evidenceValid&&model.sameProcessVerified===true&&reply.sameProcessVerified===true;
 if(!sameResources)result.issues.push('sources:resource-readings-differ-or-missing');
 if(model.evidenceValid!==true)result.issues.push('model:evidence-invalid');
 if(reply.evidenceValid!==true)result.issues.push('reply:evidence-invalid');
 if(model.model&&reply.before&&reply.after){
  try{
   const a=stamp(model.model.before.checkedAt,true),b=stamp(model.model.after.checkedAt,true),c=stamp(reply.before.snapshot.checkedAt,true),d=stamp(reply.after.snapshot.checkedAt,true);
   const start=a.ns>c.ns?a:c,end=b.ns<d.ns?b:d;
   const duration=end.ns-start.ns;
   result.windowsOverlap=duration>0n;
   result.overlap={kind:duration>0n?'positive':duration===0n?'boundary-only':'none',...(duration>=0n?{startedAt:start.text,endedAt:end.text,durationNanoseconds:duration.toString(),durationSeconds:Number(duration)/1e9}:{})};
  }catch{result.evidenceValid=false;result.issues.push('windows:invalid-time');}
 }
 if(!result.windowsOverlap)result.issues.push('windows:no-positive-intersection');
 if(!result.resourceChecksPassed)result.issues.push('resource:checks-not-passed');
 if(!result.sameProcessVerified)result.issues.push('association:process-not-verified');
 result.compatible=result.evidenceValid&&result.resourceChecksPassed&&result.sameProcessVerified&&result.windowsOverlap&&model.compatible===true&&reply.compatible===true;
 result.issues.push('acceptance:real-business-not-verified');return result;
}
export async function checkCombinedObservation({resource,model,before,after}){
 // Each tool reopens the resources. Compare both hashes to reject mid-read replacements.
 const modelResult=await checkEvidence({resource,model});
 const replyResult=await checkReplyEvidence({resource,before,after});
 return combineObservation(modelResult,replyResult);
}
export const combinedExitCode=result=>result.evidenceValid!==true?1:result.compatible===true?0:2;
export async function main(args=process.argv.slice(2)){
 try{
  const options={};
  for(let i=0;i<args.length;i+=2){const key=args[i];if(!['--resource','--model','--before','--after','--output'].includes(key)||!args[i+1]||args[i+1].startsWith('--')||Object.hasOwn(options,key.slice(2)))throw Error('invalid-arguments');options[key.slice(2)]=args[i+1];}
  if(!options.resource||!options.model||!options.before||!options.after)throw Error('invalid-arguments');
  const result=await checkCombinedObservation(options);if(options.output)await writeEvidence(options.output,result);process.stdout.write(JSON.stringify(result,null,2)+'\n');return combinedExitCode(result);
 }catch{process.stderr.write('联合观察核对失败：请检查参数、资料读取权限，以及输出文件是否尚不存在。\n');return 1;}
}
if(process.argv[1]&&import.meta.url===pathToFileURL(path.resolve(process.argv[1])).href)process.exitCode=await main();
