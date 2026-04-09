<?php
/**  Histrix Error.
 * Get and report error data and send home
 *
 */

class Histrix_Error extends Histrix{

    
    function __construct($message='') {
        parent::__construct();

        $this->debugData = $this->getDebugData($message);
        
    }

    /**
     * Get Debug Data
     */
    function getDebugData($message) {

        $debugData['message']   = $message;        
        $debugData['prefs']     = (isset($_SESSION['user_preferences']))?$_SESSION['user_preferences']:'';
        //$debugData['last_sql']  = $_SESSION['last_select_sql'];
        $debugData['last_update_sql']  = (isset($_SESSION['last_update_sql']))?$_SESSION['last_update_sql']:'';
        
        $debugData['time']      = processing_time();

        return $debugData;
    }


    function send(){
        // if there is mail
     //   if ($_SESSION['smtpServer'] != ''){
    //        $this->sendMail();
       /* }
        else {
            $this->sendBugzilla();
        }
        */

    }

    function sendBugzilla(){
        require_once '../lib/bugzillaphp/BugzillaConnector.php';
        $bugzilla = new BugzillaConnector('http://intranet.estudiogenus.com:8080/bugzilla/', $_SESSION['bugzillaCookies']);
        $bugzilla->username = "---@estudiogenus.com";
        $bugzilla->password = "---";
        $userId = $bugzilla->login($bugzilla->username,$bugzilla->password,true);

        $_SESSION['bugzillaCookies'] = $bugzilla->getCookies();

        $bugzilla->getBugzillaConfig();

       // SearchForBugs($bugzilla);

        $params = new BugzillaBug();
        $params->assigned_to = array('luis@estudiogenus.com');
        $params->bug_severity = 'normal';
        $params->bug_status = 'NEW';
        $params->priority = 'P2';
        $params->product = 'Histrix';
        $params->reporter = array('luis@estudiogenus.com');
        $params->component = 'Subproduct';
        $params->version = 'unspecified';
        $params->rep_platform = 'web';
        $params->op_sys = 'other';
        $params->short_desc = 'Histrix Transaction Bug';
                $content = print_r($this->debugData, true);
        $params->comment = 'Error en histrix';
        $params->bug_file_loc = 'http://intranet.estudiogenus.com';
        $bugzilla->createBug($params);

        print "Bug: ".$params->bug_id;



    }

    // send by mail
    function sendMail(){

        $tomail = 'luis@estudiogenus.com';

        $content = '<pre>'.print_r($this->debugData, true).'</pre>';

//        $content = $this->debugData['message'];
//        $content = 'test';
        $db = $_SESSION["db"];
        $database = Cache::getCache('datosbase'.$db);

        require_once '../lib/Swift4/swift_required.php';

            /*
        $smtpServer= $_SESSION['smtpServer'];
        $email=      $_SESSION['email'];
        $userName=   $_SESSION['userName'];*/

        $smtpServer= 'smtp.gmail.com';
        $email=      'luis@estudiogenus.com';
        $userName=   'luis@estudiogenus.com';


        //echo $email;
        $user= '---@estudiogenus.com';
        $pass= '----';

        //Create the Transport the call setUsername() and setPassword()
        $transport = Swift_SmtpTransport::newInstance($smtpServer, 465, 'ssl')//
         ->setUsername($user)
         ->setPassword($pass);
        //Create the Mailer using your created Transport
        $mailer = Swift_Mailer::newInstance($transport);
        $date = date('d/m/Y');

        //Create a message

        $message = Swift_Message::newInstance("Histrix Error ".$db);
        $message->setFrom($email)->setTo($tomail);

        $message->setBody($content, 'text/html');
        //Add alternative parts with addPart()
        $message->addPart($content, 'text/plain');

        //Pasa a variable name to the send() method
        if (!$mailer->send($message, $failures)){
            echo "Failures:";
            print_r($failures);
        }
        else {
            return;
        }

        /*
        // COPY TO SENT FOLDER
        $server= $database->imapServer;
        $serverImap = "{".$server.":143/imap/notls}";
 
            if ($server == 'imap.gmail.com') {
                $serverImap = '{imap.gmail.com:993/ssl}';
            $stream = imap_open( $serverImap."[GMail]/Enviados", $user, $pass);
        imap_append($stream, "{imap.example.org}[GMail]/Enviados", $message->getHeaders()."\r\n".
                                       $message->getTo()."\r\n".
                                       $message->getSubject()."\r\n".
                                       $message->getBody()."\r\n");
           }                 
    
            */


    }



}

?>