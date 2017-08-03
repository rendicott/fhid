import redis

r = redis.StrictRedis(host='localhost', port=6379, db=0)

for key in r.scan_iter():
    print key 

print r.get('5d1af9d5-7f1f-442e-b250-c5692d4b5065')
