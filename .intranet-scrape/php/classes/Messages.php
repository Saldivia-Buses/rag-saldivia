<?php

/**
 * Message Manipulation
 * 2010-05-28
 */
class Messages {

    public function __construct($user) {
        $this->user = $user;
        // Conditions
        // 0  - Unread
        // 1  - Readed
        // 2  - Remainder

        $this->readCondition = 0;
        $this->ExtraConditions[] = ' HTXMESSAGES.idEvento = \'\' ';
        
        
    }

    public function deleteMessage($id) {

        // mark as read
        $strSQL = 'update HTXMESSAGES set leido =1 where HTXMESSAGES.para = \'' . $this->user . '\'  and HTXMESSAGES.idMensaje = ' . $id;
        consulta($strSQL, null, 'nolog');
    }

    function sqlQuery() {
        $strSQL = 'select * from HTXMESSAGES left join HTXUSERS on HTXUSERS.login = HTXMESSAGES.de where ';
        $strSQL .= ' HTXMESSAGES.para = \''.$this->user.'\' and ';
        $strSQL .= 'TIMESTAMP(fecha, hora) <= now() and'; // ignore future notifications

        if (isset($this->ExtraConditions)) {
            foreach ($this->ExtraConditions as $num => $condition) {
                $strSQL .= $condition.' and ';
            }
        }
        $strSQL .= ' HTXMESSAGES.leido = '.$this->readCondition;
        $strSQL .= ' order by HTXMESSAGES.fecha , hora DESC';

        return $strSQL;
    }

    public function getMessages() {

        $strSQL = $this->sqlQuery();
        $rsPrivado = consulta($strSQL, null, 'nolog');

        $this->numfilas = _num_rows($rsPrivado);
        $script = '';
        if ($this->numfilas > 0) {
            while ($row = _fetch_array($rsPrivado)) {
                $this->processResults($row);
            }

            $this->refreshTrayIcon();
        }

    }

    public function setTitle($title){
        if ($this->numfilas > 0) {
            $messageString = '(' . $this->numfilas . ') ' . $title . '          ';
//            $this->javascriptCode[] = 'Histrix.message="' . $messageString . '";';
            //$this->javascriptCode[] = 'if (refreshTitle != false) Histrix.titler();';
            $this->javascriptCode[] = 'Histrix.titler(\''.$messageString.'\');';
        } else {
            // No messages
            $this->javascriptCode[] = 'Histrix.message=""';
            $this->javascriptCode[] = 'Histrix.titler(\'histrix\')';

        }
        $this->javascriptCode[] = "Histrix.osTray({count:".$this->numfilas."}); ";        
    }
    protected function processResults($row) {
        $uid = $row['idMensaje'];

        // TEST one reading at a time
        $xmllector  = 'message_recive_single.xml';
        $idxmllector 	= str_replace('.','_',$xmllector).$uid;
        $Descripcion = $row['Nombre'].' '.$row['apellido'].' - '.$row['hora'];
        $texto = $row['mensaje'];        
        $dir = 'histrix/mensajeria';
        $dirXML .=$dir.'/';
        $textoEsc =  str_replace(["\r","\n", '"'],['<br>','<br>',' ' ],addslashes($texto)) ;
        $DescripcionEsc = addslashes($Descripcion);           
        
        $this->javascriptCode[] = "Histrix.loadInnerXML('DIV".$idxmllector."', 'histrixLoader.php?xml=".$xmllector."&dir=".$dir."&leido=0&idMensaje=$uid&xmlOrig=$uid&idxmluid=$idxmllector', {evalScripts:true, asynchronous:true}, '$Descripcion', null, null, {checkDuplicate:true, maximize:false,width:\"480px\", height:\"270px\"});";
	$this->javascriptCode[] = "Histrix.desktopNotify({title:'$DescripcionEsc', text:\"$textoEsc\"}, '$uid')";            

    }


    function refreshTrayIcon($num='') {
        if ($this->trayId != ''){
            $this->javascriptCode[] = "$('#" . $this->trayId . "').remove();";
//            $this->javascriptCode[] = "var TrayIcon = jQuery('<div class=\"" . $this->className . "\" id=\"" . $this->trayId . "\" style=\"background-image:url(\'". $this->image ."\');\" >$num<img valign=\"middle\" title=\"" . $this->title . " ($num)\" src=\"" . $this->image . "\"></div>').click(function(){  $('" . $this->toggleTarget . "').toggleClass('" . $this->toggleClass . "')});";
            $this->javascriptCode[] = "var TrayIcon = jQuery('<div class=\"" . $this->className . "\" id=\"" . $this->trayId . "\" style=\"background-image:url(\'". $this->image ."\'); width:18px;color:white;text-align:left;font-size:9px;padding:0px;\" title=\"" . $this->title . " ($num)\" >$num</div>').click(function(){  $('" . $this->toggleTarget . "').toggleClass('" . $this->toggleClass . "')});";
            $this->javascriptCode[] = "$('#Histrix_Tray').prepend(TrayIcon); ";
        }
    }

    public function getScripts() {
	if (isset($this->javascriptCode))
            return Html::scriptTag($this->javascriptCode);
    }

}

?>
