<?php
/**
 * Notifications Widgets
 * 2010-05-28
 */



class Notifications extends Messages {



    public function __construct($user) {
        $this->user = $user;
        $this->readCondition = 0;
        $this->ExtraConditions[] = ' HTXMESSAGES.idEvento != \'\' ';

        $this->trayId       = 'notifToggle';
        $this->className   = 'traybutton';

        $this->image        = '../img/notif.png';
        $this->title        = 'Mostrar/Ocultar Notificaciones' ;
        $this->toggleClass  = 'background';
        $this->toggleTarget = '#notifications';
    }


    public function getMessages() {

        $strSQL = $this->sqlQuery();
        $rsPrivado = consulta($strSQL, null, 'nolog');

        $numfilas  = _num_rows($rsPrivado);
        $script  = '';
        if ($numfilas > 0) {
            while ($row = _fetch_array($rsPrivado)) {
                $uid = 'Notification_'.$row['idMensaje'];
                $strSQLlog = 'select * from HTXLOG where HTXLOG.idEvento = "'.$row['idEvento'].'"';
                $rslog = consulta($strSQLlog, null, 'nolog');
                $num  = _num_rows($rslog);
                $log = '';
                $cant = 0;
                if ($num > 0) {
                    $log = '<br>';
		    $log .= '<table class="changes">';
                    while ($row2 = _fetch_array($rslog)) {
                        
			if ($row2['campo'] != '') {
                            $cant++;
			    $log .='<tr>';
                            $log .=  '<th>'.$row2['campo'].': '.'</th>';
                            $log .=  '<td>'.$row2['valorNuevoTxt'].'</td>';
      
			    $log .='</tr>';
 	                 }
                        
			$clave = $row2['obs'];
                        $tipo  = $row2['tipo'];

                        if ($row2['xmlprog'] != '')
                            $program = $row2['xmlprog'];
                        if ($row2['xmldir'] != '')
                            $xmldir = $row2['xmldir'];
                    }
		    $log .= '</table>';
                }
                if ($cant == 0) $log = '';

                $date = Types::formatDate($row['fecha'], 'd/m/y');
                $user = $row['Nombre'].' '.$row['apellido'];

                $header = '<table><tr><td class="name">'.'<img   src="../img/mail_generic.png" />  '.$user.'</td><td class="date">'.$date.' '.substr($row['hora'], 0,5).'</td></tr></table>';

                $key = $clave;

                $notification = new Notification($uid, $header, addslashes($message), 0);

                //
                // unserialize previously stored Notification
                //
                if ($row['notification'] != ''){
                    $notification = unserialize($row['notification']);
		    $classname = get_class($notification);
		    
		    if ($classname != '__PHP_Incomplete_Class'){
            	        $notification->id    = $uid;
        	        $notification->title = $header;
    	                $notification->log   = $log;
	                $notification->date  = $date;
                        if ($tipo == 'INSERT'){
                	        $notification->obs = $clave;
        	        } else {
    	                    $notification->key = $clave;
	                }

                	    $notification->buildText();
		    }
                }

	        if ($classname != '__PHP_Incomplete_Class'){


            	    $notification->onClick( "$.post('sysmon.php', {'deleteNotification':". $row['idMensaje'] ."});" ) ;

            	    echo $notification->output();
		}
            }

            // Set  Icon
            $this->refreshTrayIcon($numfilas);

        }
        else {
            $this->javascripAejecutar[]="$('#".$this->trayId."').remove();";
        }
    }


}
?>
