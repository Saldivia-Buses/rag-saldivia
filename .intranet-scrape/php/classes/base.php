<?php
// Objeto con bases
class base {

	var $id;
	var $dsn;
	var $base;
	var $tipo;
	var $xmlPath;

	var $driver;
	var $user;
	var $password;
	
	function base($id){
		$this->id = $id;
            $this->tipo = DBTYPE;
	    $this->lang = LANG;
	    $this->driver = 'mysql';
	}
}

?>