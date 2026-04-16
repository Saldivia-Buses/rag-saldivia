<?php
/**
 * Cache methodologie
 */
class Cache
{
    public static function getCache($key)
    {
	$connection = false;
        if (function_exists('memcache_connect')) {
            $memcache = new Memcache();
            $connection = @$memcache->connect('localhost',11211);

            if ($connection) {
                $Data = $memcache->get($key);

                return $Data;
            }
        }
        if (!$connection) {
            if (isset($_SESSION[$key])) {

                $Data = $_SESSION[$key];

                return $Data;
            }
        }

        return false;
    }

    public static function setCache($key, $value, $expiration=0)
    {
    // use memcache
        if (function_exists('memcache_connect')) {
            $memcache = new Memcache();
            $connection = @$memcache->connect('localhost',11211);
            if ($connection) {

                $memcache->set($key,$value,0,$expiration);

                return;
            }
        }

        // Session method
        $_SESSION[$key]= $value;

    }

}
