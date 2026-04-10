<?php

// Retrive User Info
// keep alive connection
// Check update Libraries.

include ('./autoload.php');
include ('./sessionCheck.php');
include_once('../funciones/conexion.php');


$xmlPath = $datosbase->xmlPath;
$dirXML = '../database/'.$xmlPath.'/xml/';
$fotoPath = 'fotos/usuarios/';
$fullPath = $dirXML.$fotoPath;

$userLogin= $_POST['user'];

$user = new usuario($userLogin);
$user->email    = $_POST['email'];
$user->name     = $_POST['name'];
$user->foto     = $_POST['foto'];
$user->profile  = $_POST['profile'];
$user->interno  = $_POST['interno'];
$user->telefono = $_POST['telefono'];
//$user->getData(); PAINFULY SLOW


?>
    <div class="userOptions">Información</div>
   <div class="userData">
        <?php
            echo '<div><b>'.$user->name.'</b></div>';
            if ($user->email != ''){
                echo $user->email;
            }
            if ($user->interno != 0){
                echo '<div>Interno: ';
                echo $user->interno;
                echo '</div>';
            }
            if ($user->telefono != 0){
                echo '<div>Tel&eacute;fono: ';
                echo $user->telefono;
                echo '</div>';
            }

            if ($user->profile != ''){
                echo '<div>Perfil: ';
                echo $user->profile;
                echo '</div>';
            }
        ?>
   </div>
   <div class="userPhoto">
   <?php
        if ($user->foto != ''){
            $archivo 	= new Archivo($user->foto, $fullPath, '');
            echo $archivo->thumb(100, 100,null);
        }
   ?>
   </div>
