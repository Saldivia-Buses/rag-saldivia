<?php
/*
 * Created on 19/04/2008
 *
 * To change the template for this generated file go to
 * Window - Preferences - PHPeclipse - PHP - Code Templates
 * Registry Class
 * inpired by http://www.phpit.net/article/using-globals-php/3/
 */
// NOT USED YET!!!
 //TODO: implement as configuration method
 Class Registry {
    var $_objects = array();

    function set($name, &$object) {
            $this->_objects[$name] =& $object;
    }

    function &get($name) {
            return $this->_objects[$name];
    }

    static function &getInstance() {
    	static $me;

        if (is_object($me) == true) {
                return $me;
        }

        $me = new Registry;
        return $me;
	}

}
?>
