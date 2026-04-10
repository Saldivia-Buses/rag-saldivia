<?php
/*
 * Created on 10/08/2006
 * borro las variables de session asociadas a esta solapa
 * para ahorrar memoria (y ganar velocidad?)
 */

require ("./autoload.php");
require ("../funciones/utiles.php");
 include ("./sessionCheck.php");
 
 //$winid     = $_REQUEST['__winid'];
  $var1     = $_GET['xmldatos'];
 $xmlOrig   = $_GET['xmlOrig'];
 $inst      = $_REQUEST['instance'];

$instances = $_REQUEST['instances'];

$instances[$inst] = $inst;

if ($inst == 'undefined') {
return;
} 

Histrix_Functions::removeInstances($instances, 'all');

 
 // Delete all
 //$_GET['all'] == 'true' ||
 //
 if (isset($clearSessionData) && $clearSessionData == true){
     unset($_SESSION['instances']);
//     unset($_SESSION['xml'.$__winid]);
 
 }

if (isset($_REQUEST['session_close']) && $_REQUEST['session_close'] == 1){
    Histrix_Functions::sessionClose();
}
 ?>
