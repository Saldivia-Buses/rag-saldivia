<?php
/**
 * Unique ID
 *
 * @author luis
 */
class UID {

    public static function getUID($prefix ='', $param=null) {
        return $prefix . time() . substr(md5(microtime()), 0, rand(5, 12));
    }

}

?>
