"""Verify the running service through HTTP, including failure responses."""
import json
import os
import sys
import time
import urllib.error
import urllib.request

base=sys.argv[1]
def request(method,path,body=None,status=200):
    payload=json.dumps(body).encode() if body is not None else None
    req=urllib.request.Request(base+path,data=payload,method=method,headers={'Content-Type':'application/json'})
    try: response=urllib.request.urlopen(req,timeout=5)
    except urllib.error.HTTPError as error: response=error
    with response:
        data=response.read()
        assert response.status==status,(method,path,response.status,data)
    result=json.loads(data) if data else None
    print(method,path,status,json.dumps(result),flush=True)
    return result

for attempt in range(30):
    try:
        request('GET','/healthz')
        break
    except (OSError,AssertionError):
        if attempt==29: raise
        time.sleep(1)
info=request('GET','/')
if os.environ.get('EXPECTED_VERSION'):
    assert info['version']==os.environ['EXPECTED_VERSION'],info
result=request('POST','/api/plan',{'replicas':7,'batch_size':3,'seconds_per_batch':20})
assert result=={'batches':3,'last_batch_size':1,'estimated_seconds':60},result
request('POST','/api/plan',{'replicas':0,'batch_size':1,'seconds_per_batch':20},status=422)
print('PASS: plan calculation and validation through the Kubernetes Service')
