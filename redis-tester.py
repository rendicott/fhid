import redis
import sys 


r = redis.StrictRedis(host='localhost', port=6379, db=0)

#for key in r.scan_iter():
#    print key 

print r.get(sys.argv[1])
