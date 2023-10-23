# mocache
mocache implements the CLOCK-Pro caching algorithm.
CLOCK-Pro is a patent-free alternative to the Adaptive Replacement Cache, https://en.wikipedia.org/wiki/Adaptive_replacement_cache. It is an approximation of LIRS ( https://en.wikipedia.org/wiki/LIRS_caching_algorithm ), much like the CLOCK page replacement algorithm is an approximation of LRU.
This implementation is based on the python code from https://bitbucket.org/SamiLehtinen/pyclockpro .
Slides describing the algorithm: http://fr.slideshare.net/huliang64/clockpro
The original paper: http://static.usenix.org/event/usenix05/tech/general/full_papers/jiang/jiang_html/html.html
It is MIT licensed, like the original.
