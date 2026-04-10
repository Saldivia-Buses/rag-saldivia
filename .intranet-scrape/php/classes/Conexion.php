<?php

class Conexion {

// Variable que tiene la conexion
    var $link;
    //var $linkMETA;

    var $Driver;

    function __construct($base) {
        if (isset($base->driver))
            $this->Driver = $base->driver;
		/*
         *  al crear el objeto me conecto */
        $cursor = '';
        if (defined( 'SQL_CUR_USE_DRIVER' ))
            $cursor = SQL_CUR_USE_DRIVER;

        $this->link     = $this->Conectar($base, $cursor);
        //$this->linkMETA = $this->Conectar($base, SQL_CUR_USE_ODBC);

        switch ($base->driver) {
            case "mysql":
                $_SESSION['ffecha']= 'yyyy-mm-dd';
                break;
            case "firebird":
                $_SESSION['ffecha']= 'yyyy-mm-dd';
                break;
        }

    }

    /**
     * Database connection
     * @param string $base
     * @param int $cursor
     * @return <type>
     */
    function Conectar($base, $cursor='') {
        $enlace = $this->link;

 /*       if ($cursor == SQL_CUR_USE_ODBC)
            $enlace = $this->linkMETA; */

        if (!($enlace)) {

            $odbcUso 	=  (isset($base->dsn) && $base->dsn != '')?$base->dsn:'';

            $user 	    = $base->user;
            $password 	= $base->password;
            $host 	=  (isset($base->host) && $base->host != '')?$base->host:'localhost';
            $port       =  (isset($base->port) && $base->port != '' )?$base->port:3306;
            // Conexion propiamente dicha
            try {
                
                $enlace = _connect($odbcUso, $user, $password, $cursor, $base->base, $host, $port);

            } catch (Exception $e) {
                die($e);
            }
            $this->Driver       = $base->driver;
            $_SESSION['Driver'] = $base->driver;
        }
        return $enlace;
    }
}
?>