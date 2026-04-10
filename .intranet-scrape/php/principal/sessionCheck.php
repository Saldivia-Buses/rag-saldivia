<?php
/*
 * Created on 18/09/2008
 * Check if Session has started
 */
if (!isset($_SESSION)) {
    session_start();
}
if ($_SESSION['validado'] !== true || $_SESSION['usuario'] == '') {

	echo Histrix_Functions::sessionClose();
	
}
if (isset($_GET['DAT']) && isset($_SESSION['DAT']) && $_GET['DAT'] == $_SESSION['DAT']) {
	$DirectAccess = true;
}
/*
if($_SERVER['HTTP_X_REQUESTED_WITH'] !== 'XMLHttpRequest' && $DirectAccess !== true) {
	//Request identified as ajax request
	echo 'Direct access not Allowed!';
	die();
}
*/


?>