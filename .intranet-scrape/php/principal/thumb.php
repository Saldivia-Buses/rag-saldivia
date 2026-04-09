<?php
///////////////////////////////
//   Select Graphic Library  //
///////////////////////////////

$DirectAccess = true; //allow direct Access for this file
include ("./sessionCheck.php");

if (class_exists('Imagick')){

    include('thumb_ima.php');
    }
else {
    include('thumb_gd.php');
}
?>