<?php
/*
 * Inicio de Session del sistema
 * Registra las variables de session del usuario y base y valida sal usuario
 *
 */
// include ('./autoload.php');

//include ('./sessionCheck.php');
session_start();

// Inicio de session remota
$db = $_GET["db"];
$_SESSION['db'] = $db;
$_SESSION['usuario'] = $_GET['user'];
$_SESSION['pass'] 	 = $_GET["p"];
$_SESSION['horaIngreso'] = time();


//include_once ('../funciones/config.php');

$config = new config('config.xml', '../database/', $db);
$nom_empresa = $config->nom_empresa;

include ("../funciones/conexion.php");

$db = $_GET["db"];
$datosbase = $config->bases[$db];

Cache::setCache('datosbase'.$db, $datosbase);

//$_SESSION['datosbase'] = $datosbase;



	$cant = 0;
	if ($_SESSION['usuario']!=''){

		$sql = 'select * from HTXUSERS where login='."'".$_SESSION['usuario']."'".' and pass= '."'".$_GET["p"]."'";
		$res = consulta($sql);
		while ($row = _fetch_array($res)) {
			$cant ++;
			$_SESSION['id_usuario'] = $row['Id_usuario'];
		}
	}
	else {
		$mensaje='Nombre de Usuario no puede estar vacio';

	}

	/* Si la conexion con la tabla de Usuarios no devuelve datos NO Valido la session */
if ($cant != 0) {
	$_SESSION['validado'] = true;
} else {
	$_SESSION['validado'] = false;
	echo "<b>NOMBRE DE USUARIO O CONTRASEÑA INCORRECTA</b>";
	echo "<br/><b>".$mensaje."</b>";
	echo '<br/>';
}
?>
