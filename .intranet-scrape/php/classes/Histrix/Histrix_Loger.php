<?php
/* 
 * Loger Class 2009-05-22
 * Uses HTXLOG Table
 * @author Luis M. Melgratti
 */

class Histrix_Loger {
    
    function __construct($xml, $title ='', $wherelog ='', $updatelog='') {
        $this->xml= $xml;
        if ($title == '') $title = $xml;
        $this->title= strip_tags($title);
        $this->wherelog = $wherelog;
        $this->updatelog = $updatelog;
        $this->logcount = 0;
    }

    function log($tipo) {
        $idEvento=uniqid($this->xml.'|');
        $_SESSION['_idEvento'] = $idEvento;
        if (is_array($this->wherelog)){
            $obs = $this->wherelog['__ref__'].' ';
            foreach($this->wherelog as $campoWhere => $valorWhere) {
                if ($campoWhere != '__ref__')
                  $obs .= $campoWhere.': '.$valorWhere.', ';
            }

        } else {
            $obs = $this->wherelog;
        }

        switch($tipo) {
            case 'insert':
                $sqlLOG  = 'INSERT IGNORE INTO HTXLOG (fecha, xmlprog, xmldir, nombre, tipo, usuario, obs, idEvento) ';
                $sqlLOG .= 'values(now(), "'.$this->xml.'", "'.$this->dir.'","'.$this->title.'", "INSERT", "'.$_SESSION['usuario'].'", "'. addslashes( $obs ).'", "'.$idEvento.'")';
                $this->logcount++;
                $this->save($sqlLOG);
                break;
            case 'update':
                if ($this->updatelog != '') {

                    foreach($this->updatelog as $campoUp => $valoresUp) {
                        if ($valoresUp[0] != $valoresUp[1]) {
                            $valoresUp[0] = addslashes($valoresUp[0]);
                            $valoresUp[1] = addslashes($valoresUp[1]);
                            
                            $valoresUp[2] = addslashes($valoresUp[2]);
                            $sqlLOG  = 'INSERT IGNORE INTO HTXLOG (fecha, xmlprog, xmldir,nombre, tipo, usuario, obs, campo, valorViejo, valorNuevo, valorNuevoTxt, idEvento) ';
                            $sqlLOG .= 'values (now(), "'.$this->xml.'", "'.$this->dir.'", "'.$this->title.'", "UPDATE", "'.$_SESSION['usuario'].'", "'.addslashes($obs).'" , "'.$campoUp.'", "'.$valoresUp[0].'", "'.$valoresUp[1].'", "'.$valoresUp[2].'", "'.$idEvento.'")';
                            $this->logcount++;
                            $this->save($sqlLOG);
                        }
                    }
                }
                break;
            case 'delete':
                $sqlLOG  = 'INSERT IGNORE HTXLOG (fecha, xmlprog, xmldir, nombre, tipo, usuario, obs, idEvento) ';
                $sqlLOG .= 'values(now(), "'.$this->xml.'", "'.$this->dirx.'", "'.$this->title.'", "DELETE", "'.$_SESSION['usuario'].'", "'.addslashes($obs).'" , "'.$idEvento.'")';
                $this->save($sqlLOG);

                $this->logcount++;
                break;

        }

    }

    public function save($sqlLOG){
       updateSQL($sqlLOG, 'insert',  $this->xml);
    }

}
?>
