<?php
/* 
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */


class databaseError{

    function __construct($error, $errorNumber){
    	//default message
    	$this->errorMessage = 'Error de Update : '.$error."\n";
    }

}


?>
