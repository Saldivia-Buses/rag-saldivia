<?php
/**
 * Notification Widgets
 * 2010-05-28
 */



class MessageRemainder extends Messages {




    public function __construct($user) {
        $this->user = $user;
        $this->readCondition = 2;                                    
        $this->trayId = 'postitToggle';
        $this->className   = 'traybutton';
        $this->image        = '../img/note.png';
        $this->title        = 'Mostrar/Ocultar Notas' ;
        $this->toggleClass  = 'foreground';
        $this->toggleTarget = '.postit';
//        $this->toggleTarget = '.postit';

    }


    public function getMessages() {


        $strSQL = $this->sqlQuery();
        $rsPrivado = consulta($strSQL, null, 'nolog');

        $numfilas  = _num_rows($rsPrivado);
        $script  = '';
        if ($numfilas > 0) {
            while ($row = _fetch_array($rsPrivado)) {
                $uid = 'postit_'.$row['idMensaje'];

                $date = Types::formatDate($row['fecha'], 'd/m/y');
                $user = $row['Nombre'].' '.$row['apellido'];

                $header = '<table width="100%"><tr><td class="name"><img   src="../img/question.gif" />  '.$user.'</td><td class="date" >'.$date.' '.substr($row['hora'], 0,5).'</td></tr></table>';

                $key = $clave;
                $message =  '<span>'.$row['mensaje'].'</span>';
                $message = nl2br(addslashes($message));
                $message = trim( preg_replace( '/\s+/', ' ', $message) );

                $notification = new Postit($uid, $header, $message, 0);
                $notification->onClick( "Histrix.sysmon({'deletePostit':". $row['idMensaje'] ."});" ) ;

                echo $notification->output();
            }

            // Set Notes Icon
            $this->refreshTrayIcon();


        }
        else {
            $this->javascripAejecutar[]="$('#".$this->trayId."').remove();";
        }

    }

}
?>
