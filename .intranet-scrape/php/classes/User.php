<?php
/** Clase usuario.
 * Almaceno las caracteristicas del usuario Actual
 * Nivel de acceso y propiedades
 */

class User extends Histrix{

    /* NOMBRE DE LA TABLA QUE ALMACENA LOS USUARIOS */
    var $TABLAUSUARIOS;

    var $nombreUsuario;
    var $password;
    var $nivelAceso;
    var $habilitado;

    var $preferences;

    function __construct( $user, $path ='' ) {
        parent::__construct();

        $fotoPath = 'fotos/usuarios/';

        $this->imagePath = $path.$fotoPath;


        $this->login= $user;
        $this->TABLAUSUARIOS = 'HTXUSERS';
        
        //		$this->habilitado = $this->checkUserPass();
        if (!isset($_SESSION['user_preferences']))
            $this->getPreferences();

	$this->getData();
        $this->updateAccessTime();

    }

    /**
     * update Access time log
     */
    function updateAccessTime(){
        $timestamp = time();
        $this->checkAccessTime();
        $sqlUSER  = 'update HTXUSERS ultimolog set ultimolog= CURRENT_TIMESTAMP()  where login =\''.$this->login.'\'';
        consulta($sqlUSER, null, 'nolog');
    }

    /**
     * TIME CHECK
     */

    function checkAccessTime(){

        if (isset($_SESSION['minHour']) && $_SESSION['minHour'] != '' && $_SESSION['maxHour'] != '') {

            $currentTime= time();

            $minTime = strtotime($_SESSION['minHour']);
            $maxTime = strtotime($_SESSION['maxHour']);

            if ($currentTime <= $maxTime && $currentTime >= $minTime) {
            // all ok
            // echo 'ok??';
            }
            else {
                $cant= 0;
                $script[] = 'Histrix.destroy("EXPIRO LA SESSION");';
                session_destroy();
                echo Html::scriptTag($script);
                die ('TIME ACCES CONTROL FAIL');
            }
        }

    }

    function getData() {
        $strSQL  = 'select * from '.$this->TABLAUSUARIOS .' left join HTXPROFILES on HTXPROFILES.Id_perfil = '.$this->TABLAUSUARIOS .'.Id_perfil' ;
        $strSQL .= ' where login = "'. addslashes($this->login) . '"';

        $rs = consulta($strSQL);
        while ($row = _fetch_array($rs)) {
        
//    	    if ( isset($row['readonly'])){
		$_SESSION['readonly'] = $row['readonly'];
		
//    	    }
        
            foreach($row as $key => $value)
                $this->{$key} = $value;
        }
    }

    function checkUserPass () {

        $strSQL  = 'select * from '.$this->TABLAUSUARIOS .' ';
        $strSQL .= 'where _Userid = "'. $this->login . '"';

        $rs = consulta($strSQL);
        //_result_all($rs);
        return true;
    }
    
    // Get User Preferences
    function getPreferences() {
        $sql = 'select * from HTXPREFS where login="'.$this->login.'"';
        $rs = consulta($sql,null,'nofatal');
        while ($row = _fetch_array($rs)) {
            foreach($row as $key => $value)
                $this->preferences[$key]=$value;
        }
    }

    function getUserList() {

        $usersStatus = Cache::getCache('usersStatus' );

        $mailico = '<img style="float:right;" src="../img/mail_generic.png" class="send_msg" />';
        $sqlUSER  = 'select * , HTXPROFILES.nombre as "perfil", (SELECT CAST(count(*) as SIGNED) from HTXMESSAGES where HTXMESSAGES.para = login and HTXMESSAGES.de = \''.$_SESSION['usuario'].'\' and idEvento = \'\'and HTXMESSAGES.leido = 0) as "noleidos"' .
            ', (TIMEDIFF(current_timestamp(),ultimolog) <  \'00:02:00\') as "tiempo" from HTXUSERS left join HTXPROFILES on HTXPROFILES.id_perfil = HTXUSERS.id_perfil  where baja = 0 order by apellido, HTXUSERS.Nombre';
        $rs = consulta($sqlUSER, null, 'nolog');
        if ($rs) {
            $tot = 0;
            $i = 0;
            while ($row = _fetch_array($rs)) {
                $dif = $row['tiempo'];
                $Nombre = urlencode($row['Nombre'].' '.$row['apellido']);
                $noleidos = $row['noleidos'];
                $strNoleidos = '';
                $style='';

                $username = $row['login'];
                $status = isset($usersStatus[$username])?$usersStatus[$username]:'';
                //$class = '';
                if ($status == 'true' || !$dif){
                    $style='color:#bbb;font-weight:bold;';
                }

                if ($noleidos > 0) {
                    $strNoleidos = ' ('.$noleidos.')';
                    $style='color:#ff7200;font-weight:bold;';
                }
                $foto = "thumb.php?url=".urlencode($this->imagePath.$row['foto']).'&ancho=100&alto=100';
                
                $user = '<li profile="'.$row['perfil'].'" '.
                    'title="'.$row['email'].'" '.
                    'email="'.$row['email'].'" '.
                    'foto="'.$foto.'"'.
                    'login="'.$row['login'].'" '.
                    'telefono="'.$row['telefono'].'" '.
                    'interno="'.$row['interno'].'" '.
                    'style="'.$style.'">'.
                    $row['apellido'].', '.
                    $row['Nombre'].
                    $strNoleidos;

                if ($dif) {
                    $i++;
                    $online[] = $user.$mailico.'</li>';
                }
                else {
                    $offline[] = $user.'</li>';
                }
                $tot++;
            }
            $userlist = '<ul style="margin:0px;">';
            $userlist .= '<li class="usrgrp"><div onclick="$(\'#_usronline\').toggle();">'.
                $this->i18n['connected'].' <span style="color:green;">('.count($online).')</span></div><ul id="_usronline" class="users">';
            if (is_array($online))
                $userlist .= implode(' ',$online);
            $userlist .= '</ul></li>';
            $userlist .= '<li class="usrgrp"><div onclick="$(\'#_usroffline\').toggle();">'.
                $this->i18n['disconected'].' <span style="color:red;">('.count($offline).')</span></div><ul id="_usroffline" class="users">';
            if ($offline != '')
                $userlist .= implode(' ',$offline);
            $userlist .= '</ul></li>';
            $userlist .= '</ul>';

            $this->totalUsers = $tot;
            $this->conectedUsers = $i;

            return $userlist;

        }
    }
    /**
     * 
     * @return String
     */
    function getConectedUsers() {
        $conectedString = $this->i18n['connectedUser'];
        if ($this->conectedUsers > 1)
            $conectedString = $this->i18n['connectedUsers'];
        return $this->conectedUsers.'/'.$this->totalUsers.' '.$conectedString;

    }

}

?>