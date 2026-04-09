<?php
/*
 * Close Session del sistema
 * Unregister an kills current Session
 *
 */
//include      ("./sessionCheck.php");            // Check if a valid session exists
session_start();
session_unset();
session_destroy();
header('Location:../');
?>
