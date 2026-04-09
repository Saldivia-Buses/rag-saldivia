<?php
/**
 * NotificationMessage Class
 *
 * @author Luis M. Melgratti
 * @creation 2009-10-06
 */

class NotificationMessage {

    public function __construct($action, $program , $parameters = null) {
        $this->action = $action;
        $this->program = $program;

        $this->user = $_SESSION['usuario'];
        $this->path = $_SESSION['dirXML'];
        $this->subdir = 'histrix/usuarios';
        $this->xml  = 'notifica_getusers.xml';

        $parameters['user'] = $this->user;
        $parameters['action'] = $action;
        $parameters['program'] = $program;
        $this->parameters = $parameters;

    }

    public function save() {

          $targets = $this->getTarget();

        // TODO redo this reading an xml
        if (is_array($targets))
            foreach($targets as $n => $target) {
            // Armo mensaje
                $text = $this->action.' '.$this->program;
                if (isset($this->parameters['text']))
                    $text = $this->parameters['text'];


                $notification = new Notification($uid, $header, addslashes($message), 0);
                $this->parameters['text'] = $text;
                $notification->parameters = $this->parameters;

                $serializedNotification= serialize($notification);


                // Envio Mensajes
                $str = 'INSERT INTO HTXMESSAGES (de, para, mensaje, prioridad, fecha, hora, idEvento, notification)'.
                    ' values (\''.$this->user.'\', \''
                    .$target.'\', \''
                    .$text.'\', 1,now(), now(), "'
                    .$_SESSION['_idEvento'].'", \''.
                    $serializedNotification
                    .'\') ';

                updateSQL($str);
                $NOT_str = '';
            }
            
    }

    /**
     * Get target List
     */
    public function getTarget() {
        $file  = $this->path.$this->subdir.'/'.$this->xml;
        $state = false;
        $targetArray = '';
        if (is_readable($file)) {
        
            $xmlReader = new Histrix_XmlReader($this->path, $this->xml, null, null, $this->subdir );
  //          $xmlReader->addParameters($_GET);
  //          $xmlReader->addParameters($_POST);
            $xmlReader->serialize = false;

            $Container   = $xmlReader->getContainer();
            $strSQL      = $Container->getSelect();
            $rs = consulta($strSQL, null);
            while ($row = _fetch_array($rs)) {

                $target = $row['destino'];

                if ($target == $this->user) continue;
                $targetArray[] = $target;
            }

            $state = $targetArray;
        }
        else {
            //echo 'no';
        }
        return $state;
    }

}
?>
