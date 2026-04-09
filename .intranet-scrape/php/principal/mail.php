<?php

/* ini_set("session.cookie_lifetime", 10800);
  ini_set("session.gc_maxlifetime", 10800);
 */
set_time_limit  (5);
include ("./autoload.php");
include ("./sessionCheck.php");
//include_once ('../funciones/config.php');
include ('../funciones/utiles.php');
//$database = $_SESSION['datosbase'];

$db = $_SESSION["db"];
$database = Cache::getCache('datosbase' . $db);
if ($database === false) {
    if ($xmlpath == '')
        $xmlpath = '../database/';
    $config = new config('config.xml', $xmlpath, $db);
    $database = $config->bases[$db];
    Cache::setCache('datosbase' . $db, $database);
}

$registry = &Registry::getInstance();
$i18n   = $registry->get('i18n');

$server = $database->imapServer;
$emailProgram = (isset($database->emailProgram)) ? $database->emailProgram : 'webmail_at6.php';

$user = $_SESSION['emailUser'];
$pass = $_SESSION['emailPass'];

//$_SESSION['imapServer'] = $database->imapServer;
$appName = 'imap';
$script = '';
if ($user != '' && $server != '') {

    if (function_exists('imap_errors') && $server != '') {


	    imap_timeout(IMAP_OPENTIMEOUT,5);
	    imap_timeout(IMAP_READTIMEOUT,5);

        imap_errors();
        //$serverImap = "{".$server.":143/imap/notls}";

	        $serverImap = "{" . $server . "}";
	
	        $exec1 = "xmlLoader('mail', 'url=$emailProgram', {title:'@-mail', loader:'cargahttp.php'});";
	
	        if ($server == 'imap.gmail.com') {
	            $serverImap = '{imap.gmail.com:993/ssl}';
	            $users = explode('@', $user);
	            $exec1 = "Histrix.loadExternalXML('gmail', '$emailProgram&Email={$users[0]}');";
	        }
	
	        $mailbox = imap_open($serverImap, $user, $pass, OP_READONLY);

        

//get mailboxes 
//$list = imap_list($mailbox, $serverImap, "*");
        /*
          $list = imap_getmailboxes($mailbox, $serverImap, "*");
          echo '<pre>';
          print_r($list);
          print_r( imap_status($mailbox, $serverImap.'INBOX.Sent', SA_ALL));

          echo '</pre>';
         */


        echo imap_last_error ();

        if (imap_errors ()) {
            loger($user, 'imap.error');
            loger($pass, 'imap.error');
            loger(imap_last_error(), 'imap.error');

            loger(imap_errors(), 'imap.error');
        } else {
            if ($mailbox !== false) {

                // Check messages

                $check = imap_check($mailbox);
                $noleidos = imap_status($mailbox, $serverImap, SA_UNSEEN)->unseen;
                
                $texto = 'sin correo';
                //$exec2 = "xmlLoader('mail', 'url=webmail_awm.php', {title:'@-mail', loader:'cargahttp.php'});";

                $onclick1 = 'onclick="' . $exec1 . '"';

                $estilo = ' style="color:red;cursor:pointer" ';
                $preview = '';
                if ($noleidos > 0) {

                    if ($noleidos < 20) { // only gets messages if there are few
                        $result = imap_search($mailbox, 'UNSEEN');

                        if (is_array($result)) {

                            foreach ($result as $messageNumber) {
                                $header = imap_headerinfo($mailbox, $messageNumber, 80, 80);

                                $from = htmlspecialchars((string) $header->fromaddress);
                                $subject = $header->fetchsubject;
                                $preview.= '<tr><th>' . $from . '</th><td>' . utf8_decode(imap_utf8($subject)) . "</td></tr>";
                            }
                        }
                    } else {
                        $preview.= '<tr><th>'.$i18n['newmail'].'</th><td>';
                    }

                    $imapPreview = '<table class="bubble">';
                    $imapPreview .= $preview;
                    $imapPreview .= '</table>';


                    $texto = '(' . $noleidos . ') no leidos';

                    $script = '<script type="text/javascript">';
                    $script .= '$("#bubble' . $appName . '").html("' . addslashes($imapPreview) . '");';
                    $script .= 'setTimeout( "$(\"#bubble' . $appName . '\").fadeOut(\"fast\")", 7000);';
                    $script .= 'Histrix.playSound("ding");';

                    $sintags = html_entity_decode(strip_tags($imapPreview));

                    $uid++;
                    $script .= "Histrix.desktopNotify({ title:'$texto',text:'$sintags'}, '$uid', 10000)";

                    $script .='</script>';
                }

                echo '<a class="boton" ' . $onclick1 . $estilo . ' ><img align="top" src="../img/mail_generic.png" /> ' . $texto . '</a>';
                //echo '<a class="boton" '.$onclick2.$estilo.' ><img align="top" src="../img/mimetypes/mail_generic.png" />V2</a>';

                echo $script;
                imap_close($mailbox);
            }
        }
    }
}
?>