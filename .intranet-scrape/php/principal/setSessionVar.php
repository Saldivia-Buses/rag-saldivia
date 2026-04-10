<?php
/*
 * Created on 07/11/2005
 * Luis M. Melgratti
 */
 // TODO: REMOVE THIS TOTALLY INSECURE CODE!!
 // called by javascritp tree, remove when tree get replaced
session_start();
unset($_SESSION['ARBOL']);
$_SESSION['ARBOL'] = $_GET['value'];

?>
