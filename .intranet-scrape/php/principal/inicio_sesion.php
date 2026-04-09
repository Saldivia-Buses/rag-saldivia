<?php
/*
 * Inicio de Session del sistema
 * Registra las variables de session del usuario y base y valida usuario
 *
 */
// CLOSE PREVIOUS SESSION 
session_start ();
session_unset ();
session_destroy ();

include ( '../config/config.php');
include (MAINDIR . 'autoload.php');

session_start ();

$db = (isset ( $_REQUEST["db"] )) ? $_REQUEST["db"] : '';
$_SESSION['db'] = $db;
$_SESSION['usuario'] = strtolower((isset ( $_REQUEST["usuario"] )) ? $_REQUEST["usuario"] : '');
$_SESSION['horaIngreso'] = time();
$_SESSION['imgpath'] = IMGDIR;

$_SESSION['dateFormat'] = 'd/m/Y';

$_SESSION['validado'] = false;
 $_SESSION['mobile'] = $_REQUEST['mobile'];
$config = new config ( CFGFILE, DBDIR, $db );

$nom_empresa = $config->nom_empresa;

if (isset ( $config->bases[$db] )) {
	$datosbase = $config->bases [$db];
	
	//CACHEAR
	Cache::setCache ( 'datosbase' . $db, $datosbase );
	$_SESSION ['datapath'] = $datosbase->xmlPath;

	if ($datosbase->smtpServer != '')
		$_SESSION ['smtpServer'] = $datosbase->smtpServer;
	if ($datosbase->imapServer != '')
		$_SESSION ['imapServer'] = $datosbase->imapServer;

        $_SESSION ['properties'] = $datosbase->properties;
        $_SESSION ['lang'] = $datosbase->lang;

	// d2k almaceno los datos de la base activa en la sesion para poder usarlo en los XML 
        $_SESSION [ 'datosbase' . $db] = $datosbase;
}
// Incluye el encabezado con Javascript (d2k: horrible resolver con config.php)
$inipath = '../';

include_once (FUNCDIR . 'utiles.php');
include_once (FUNCDIR . 'conexion.php');
//include      (INCDIR  . 'encab.php');

// PLUGIN LOADING
$pluginLoader = new PluginLoader();

$pluginLoader->getAvailablePlugins();
$plugins = $pluginLoader->getRegisteredPlugins();

// Hook to registered plugin

$returnedValues = PluginLoader::executePluginHooks('preSessionInit', $plugins);


// para un futuro Manejo de Datos el usuario del sistema (d2k: ????)
if (isset ( $_REQUEST['usuario'] )) {
    $user = strtolower($_REQUEST['usuario']);
    $pass = md5( $_REQUEST['pass'] );
    $login = addslashes ( $user );
}

if (isset ( $_REQUEST['md5pass'] ) ) {
    $pass = $_REQUEST['md5pass'];
}

// Prevent Header breakup during check and store information for futher use
ob_start();
$sysok = buildsystemstructure($_SESSION['datapath']);
$mensaje = ob_get_contents();
ob_end_clean();

if($sysok)
{
    if ($_SESSION['usuario'] != '') {

            $sql = "SELECT * from HTXUSERS where login='$login' and pass='$pass'";
            $res = @consulta($sql,null, 'nolog');

            if (_num_rows($res) == 0) $mensaje = MSG003;

            if (isset ( $_SESSION['id_usuario'] ))
                    if ($_SESSION['id_usuario'] == strtoupper( $_SESSION ['id_usuario'] ))
                            $mensaje = MSG004;


            while ( $row = _fetch_array ( $res ) ) {
                    if (!$row['baja'] )
                    {
                        $_SESSION['idUser']   = $row['Id_usuario'];
                        $_SESSION['profile']  = $row['Id_perfil'];
                        $_SESSION['userName'] = $row['Nombre'] . ' ' . $row['apellido'];
                        $_SESSION['userImage'] = $row['foto'] ;
                                                
                        if ($row['editor'] == 1)
                            $_SESSION ['EDITOR'] = 'editor';
                        if ($row['admin'] == 1)
                            $_SESSION ['administrator'] = $row['admin'];
                        if ($row['email'] != '')
                            $_SESSION ['email'] = $row ['email'];
                        if ($row['emailUser'] != '')
                            $_SESSION ['emailUser'] = $row ['emailUser'];
                        if ($row['emailPass'] != '')
                            $_SESSION['emailPass'] = $row ['emailPass'];

                        if (is_remote_access())
                        {
                            if (!$row ['remote'])
                            {
                                $mensaje = sprintf(MSG006,$login,$_SERVER['REMOTE_ADDR']);
                            }
                            else
                            {
                                $mensaje = MSG009;
                            }
                        }
                        else
                        {
                            $mensaje = MSG009;
                        }

                        // TIME BASED ACCESS

                        if ($row ['minHour'] != '' && $row ['maxHour'] != '')
                        {
                            $currentTime = time();

                            $minTime = strtotime ( $row ['minHour'] );
                            $maxTime = strtotime ( $row ['maxHour'] );

                            if ($currentTime <= $maxTime && $currentTime >= $minTime)
                            {
                                if (isset ( $row ['minHour'] )) $_SESSION ['minHour'] = $row ['minHour'];
                                if (isset ( $row ['maxHour'] )) $_SESSION ['maxHour'] = $row ['maxHour'];
                                $_SESSION ['validado'] = true;
                            }
                            else
                            {
                                $mensaje = MSG007;
                            }
                        }
                        else
                        {
                            $_SESSION ['validado'] = true;
                        }

                        logAccess($login,$_SERVER['REMOTE_ADDR'],$mensaje);

                        // USER INTERFASE SETTINGS
                        $sql = "SELECT * from HTXPREFS WHERE login='$user'";
                        $rs = @consulta($sql,null, 'nolog');
                        if($rs)
                            while ( $row = _fetch_array ($rs) )
                            {
                                $_SESSION['userPrefs'] = $row;
                            }

                        // PRINTERS
                        $sql = "SELECT * FROM HTXPRINTERS";
                        $rs = @consulta($sql,null, 'nolog');
                        if ($rs)
                            while ( $row = _fetch_array ($rs) )
                            {
                                $_SESSION ['PRINTERS'][$row['idPrinter']] = $row['idPrinter'];
                            }

                    }
                    else
                    {
                        // SUSPENDED USER
                        $mensaje = MSG005;
                    }

            }

    }
    else
    {

        $mensaje = MSG013;
    }
}

if ($_POST['login'] == 'true'){
    header ("Content-type: text/xml");
    echo '<?xml version="1.0" encoding="UTF-8" ?>';
    echo '<login>';
     if ($_SESSION ['validado']) {
         echo '<access login="true"/>';
     }
     else{
        session_unset ();
        session_destroy ();

        echo '<error><![CDATA['.$mensaje.']]></error>';
     }
    echo'</login>';
    die();

}
print_r($_POST);
?>
    <body>
    <div class="Pagina">
        <div class="ingreso">
            <div class="error">
                <?php

                echo(MSG001);

                printf(MSG002,$_SESSION['usuario']);
                if (!isset($prism)) $prism = '';
                if ($_SESSION ['validado']) {
                    session_write_close (); // Prevents error on first open
                    echo '<script language="JavaScript">';
                    echo '	window.location="../principal/' . $prism . '";';
                    echo "</script>";
                }
                else
                {
                    session_destroy ();
                    if ($mensaje != '') echo '<br/><div style="color:red;"><b>' . $mensaje . '</b></div>';
                }
                ?>
            </div>
        </div>
    </div>
</body>
</html>