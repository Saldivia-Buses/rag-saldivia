<?php
//include('./autoload.php');

$log = 'envio.log';
//loger('entro en el programa de envio', $log);
if ($destinatarios != '')
    foreach ($destinatarios as $n => $dest) {
        $nombres[] = $dest['nombre'];
        $mails[] = $dest['EMAIL'];
        $mailnom[$dest['EMAIL']] = $dest['nombre'];
    }

try {

    require_once '../lib/Swift4/swift_required.php';

//    include_once ('../funciones/config.php');

    $db = $_SESSION["db"];
    $datosbase = Cache::getCache('datosbase'.$db);
    if ($datosbase === false){
        if ($xmlpath == '')  $xmlpath = '../database/';
        $config = new config('config.xml', $xmlpath);
        $datosbase = $config->bases[$db];
        Cache::setCache('datosbase'.$db, $datosbase);
    }

    $url = stripslashes($direccion);
    $email = stripslashes($email);
    $nombre = stripslashes($nombre);

    //$subject     = stripslashes($subject);
    $subject = $titulo;

    $text = '<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 4.01 Transitional//EN">' . "\r\n";
    $text .= '<html lang="en">' . "\r\n";
    $text .= '<head><title>' . "\r\n";
    $text .= $subject;
    $text .= '</title>' . "\r\n";
    $text .= '</head>' . "\r\n";
    $text .= '<body>' . "\r\n";

    //$cuerpo	     = stripslashes($News->show2());

    //preg_match_all("/<img(.*)src=\"(.*?)\"(.*)>/i", $cuerpo,$matches, PREG_SET_ORDER);
    //preg_match_all("/<img(.*)src=\"(.*?)\"(.*)>/i", $cuerpo,$matches, PREG_PATTERN_ORDER);
    preg_match_all('#\ssrc="(.*?)"\s#i', $cuerpo, $matches, PREG_SET_ORDER);

    $text .= $cuerpo;
    $text .= '</body>' . "\r\n";
    $text .= '</html>' . "\r\n";
    //fclose($pp);



    //loger(print_r($matches, true));
    if ($mailnom != '') {

        $mailfrom = $datosbase->mail;
        $nomfrom  = $datosbase->nombre;

        // CONNECTION POLL
        if ($datosbase->smtpServer != '') {
            $conn = Swift_SmtpTransport::newInstance($datosbase->smtpServer, 25);

            //$conn = new Swift_Connection_SMTP($datosbase->smtpServer, 25);
            if ($datosbase->smtpUser != '') {
                $conn->setUsername($datosbase->smtpUser);
                $conn->setpassword($datosbase->smtpPassword);
            }
            $connections[] = $conn;
        }

	/*	$conn1 = new Swift_Connection_SMTP("localhost");
		$connections[] = $conn1; */
        //$connections[] = new Swift_Connection_NativeMail();

        //$swift = new Swift(new Swift_Connection_Multi($connections));
        $swift = Swift_Mailer::newInstance($conn);
        $message = Swift_Message::newInstance($subject);
          /*->setFrom(trim("de"))
            ->setTo(trim("para"))
           ->setBody("mensaje")*/

        //$message = new Swift_Message($subject);

		/* rehacer con path relativo
		foreach ($matches as $att) {
			$imagen = '/home/ni000076/public_html' . $att[1];
			$img = new Swift_Message_Image(new Swift_File($imagen));
			$src = $message->attach($img);
			$text = str_replace($att[1], $src, $text);
		}*/


        //$body = new Swift_Message_Part($text, "text/html");

        $message->setBody($text, 'text/html');
        $message->setTo($mailnom);
        $message->setFrom($mailfrom);

        /*
		$recipients = new Swift_RecipientList();
		$aenviar = 0;
		foreach ($mailnom as $MailDestino => $nom) {
			$aenviar++;
			echo $MailDestino.'--->>>'.$nom.'-';
			$dest = new Swift_Address($MailDestino, $nom);
			$recipients->addTo($dest);
//            loger($MailDestino.'--->>>'.$nom.'-', $log);
		}
        */
        //		echo 'Se enviaran: ' . $aenviar . '<br>';
        //		flush();

        $cantidad = $swift->batchSend($message);
        /*
		$batch =& new Swift_BatchMailer($swift);
		//$cantidad = $batch->send($message, $recipients, $mailfrom);
		$cantidad = $swift->batchSend($message, $recipients, $mailfrom);
		//$cantidad = $swift->Send($message, $recipients, $mailfrom);
//		echo 'Se enviaron : ' . $cantidad . ' mails';

        	$noenviados = $batch->getFailedRecipients();
		if ($noenviados != '') {
			print_r($noenviados);
		}
        */

    }
} catch (Swift_ConnectionException $e ) {
    echo "There was a problem communicating with SMTP: " . $e->getMessage();
    print_r($e);
}
catch (Swift_Message_MimeException $e ) {
    echo "There was an unexpected problem building the email:" . $e->getMessage();
}

?>