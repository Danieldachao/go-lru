# go-lru
* 由于某种需求，需要一个有两种淘汰机制的本地缓存，一种是lru，即最不常访问且大于maxsize的条目会被淘汰；另一种淘汰机制是超时，即时间大于expiration时，删除此条目。
* 超时机制的实现，参考了patrickmn的go-cache https://github.com/patrickmn/go-cache.git
* 第一版只实现了最基础的功能
