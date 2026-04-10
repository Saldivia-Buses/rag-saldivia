<?php
/*
 * Histrix heartbeat
 */
// Check Internal messages
// keep alive connection
// Check update Libraries.
// Revision: 05-28-2010
//         : 12-12-2010

include ('./autoload.php');
include_once('../funciones/conexion.php');

include ("./sessionCheck.php");

$_SESSION['lastTime'] = time();

// store last time by winid to check if tab has closed
// 
$winid = $_REQUEST['__winid'];

$_SESSION['wintime'][$winid] = time();


// check all window times to check if some windows has close

Histrix_Functions::garbageCollector();


$username= $_SESSION['usuario'];

$dirXML  = '../database/'.$datosbase->xmlPath.'/xml/';

$User = new User($username, $dirXML);

$dir = 'histrix/mensajeria';
$dirXML .=$dir.'/';

$archivoxml = 'mensajes_check.xml';
$xmllector  = 'mensajes_recive.xml';

$idxmllector 	= str_replace('.','_',$xmllector);
$checkMessages = true;
if (!is_readable($dirXML.$archivoxml)) {
    $script[] = "Histrix.showMsg('Mensajeria no Instalada');";
    $script[] = " $('#Msg').fadeOut();";
    $checkMessages = false;
}

// set idle status
if (isset($_REQUEST['idle']) && $_REQUEST['idle'] != ''){
  //Cache::setCache('idle'.$username, $_REQUEST['idle']);
  $usersStatus = Cache::getCache('usersStatus' );

  $usersStatus[$username] = $_REQUEST['idle'];
  Cache::setCache('usersStatus', $usersStatus);
}
  




$plugins = Cache::getCache('Plugins');


///////////////////////////
// Check some Javascript Library has changed
///////////////////////////
/*
$javascripts[] = '../javascript/histrix.js';
$javascripts[] = '../javascript/lang/histrix-' . $datosbase->lang . '.js';
include ('javascript.php');
if ($_SESSION['lastjs_size'] != $lastjs_size) {
        
  //  $script[] = 'Histrix.showMsg("<img src=\"../img/view-refresh.png\" align=\"middle\"/> '
  //      .htmlentities($i18n['systemUpdate'])
  //      .' <a style=\"padding:2px\" class=\"boton\" href=\"#\" onclick=\"window.location.reload(true)\" >'
  //      .htmlentities($i18n['reload'])
  //      .'</a>")';
        
}
*/
///////////////////////////////////
//      Local Messages
///////////////////////////////////

$localMessages = new Messages($username);
if (isset($_POST['deleteMessage'])) {
    $Notifications->deleteMessage($_POST['deleteMessage']);
}
$localMessages->getMessages();
$localMessages->setTitle($i18n['messages']);
echo $localMessages->getScripts();
//loger($localMessages->getScripts());


///////////////////////////////////
//      Notifications
///////////////////////////////////

$Notifications = new Notifications($username);
if (isset($_POST['deleteNotification'])) {
    $Notifications->deleteMessage($_POST['deleteNotification']);
}
$Notifications->getMessages();
echo $Notifications->getScripts();


///////////////////////////////////
//      Message Remainder
///////////////////////////////////
$remainder = new MessageRemainder($username);
if (isset($_POST['deletePostit'])) {
    $remainder->deleteMessage($_POST['deletePostit']);
}
$remainder->getMessages();
echo $remainder->getScripts();

///////////////////////////
// SHOW USERS LIST
///////////////////////////

echo $User->getUserList();
$script[] = '$("#_numberusers").html(\''.$User->getConectedusers().'\');';
$script[] = "if ( $('#ultimosprogs').css('display') != 'none'){ Histrix.toggle('#ultimosprogs'); }  ";



// Hook to registered php plugins Files
echo PluginLoader::executePluginHooks('afterSysmon', $plugins);

// Hook to registered javascript plugins Files
$scriptArray = PluginLoader::executePluginHooks('afterSysmonScripts', $plugins);

if (is_array($scriptArray)){
    foreach($scriptArray as $n => $scriptjs){
        $script = array_merge($script, $scriptjs);
    }
}

echo Html::scriptTag($script);

unset($_SESSION[$archivoxml]);

?>
