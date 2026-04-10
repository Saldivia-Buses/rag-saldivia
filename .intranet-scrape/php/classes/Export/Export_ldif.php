<?php
/* 
 * Loger Class 2013-04-28
 * Export class
 * @author Luis M. Melgratti
 */

class Export_ldif extends Export{
    

    public function prepareFile(){
        $db = $_SESSION["db"];
        $database = Cache::getCache('datosbase'.$db);
        if ($database === false) {
            $config = new config('config.xml', '../database/', $db);
            $database = $config->bases[$db];
            Cache::setCache('datosbase'.$db, $database);
        }

        $this->ldapServer = $database->ldapServer;
        $this->ldapRdn    = $database->ldapRdn;
        
    }

    public function header(){

    }

    public function footer(){

    }

    public function processData($row, $Field, $params){




        $x = $params['x'];
        $y = $params['y'];
        $this->lastY= $y;
        $Valcampo = $params['value'];
        $nomcampo = $params['fieldname'];

        if (count($Field->opcion) > 0 && $Field->TipoDato != "check" && $Field->valop != 'true') {
            $valor = $Field->opcion[$Valcampo];
            if (is_array($valor)) $Valcampo = current($valor);
        }
        $valor = $Valcampo;

        if ($this->Container->seSuma($nomcampo) && $norepeat != true) {
            $Suma[$Field->NombreCampo] += $valor;
            $Subtotal[$Field->NombreCampo] += $valor;
        }

        if ($this->Container->seAcumula($nomcampo) && $norepeat != true) {
            $valor = $Suma[$Field->NombreCampo];
            $Field->TipoDato = "numeric";
        }

        $valorxls= $valor;

        $this->totals[$x]=$Suma[$Field->NombreCampo];

        switch ($Field->TipoDato) {
            case "numeric" :
                $valorxls = $valor;
            break;
            /*
            case "date":

                $cell_format = $date_format;
                $cell_format['num_format']='dd/mm/yyyy';
                $cell_format['align']='center';
                $valorxls = xl_parse_date($valor);

            break;
            
            case "time":
                $cell_format['num_format']='HH:MM:SS';
                $cell_format['align']='center';
                $valorxls = xl_parse_time($valor);
                $columnFormat[$Field->NombreCampo]=$cell_format;
                if ($this->Container->seSuma($nomcampo) || $Field->suma == 'true')
                    $this->totals[$x]='=SUM('.xl_rowcol_to_cell(1,$x).':'.xl_rowcol_to_cell($y, $x).')';

            break;
            */
            default:
                $valorxls = utf8_decode($valor);
            break;
        }                                                                                                                                                                                                                                                                                                                                                                                                                       // Si tiene opciones de un combo


        if (is_object($valorxls)){
            $valorxls = '';
        }   

        if ($Field->ldifName != '') {
            $ldifName = $Field->ldifName;
            if ($valorxls != '') {
                $this->row .= $ldifName.': '.$valorxls ;
                $this->row .="\n";
            }
            if($ldifName=='mail') {
                $this->mail = $valorxls;
            }
        }



    }

    public function endRow(){
        if ($this->mail != '') {
            $this->output.= 'dn: mail='.$this->mail.','. $this->ldapRdn; // put in configuration file
            $this->output.="\n";
            $this->output.= 'objectClass: top';
            $this->output.="\n";
            $this->output.= 'objectClass: inetOrgPerson';
            $this->output.="\n";
            $this->output.= 'objectClass: mozillaAbPersonAlpha';
            $this->output.="\n";
            $this->output.= $this->row;
            $this->output.="\n";
        }
        $this->row = '';
    }

    public function sendHeaders(){
        header ("Content-type: text/ldif");
        header("Content-Disposition:attachment; filename=\"".$this->filename.".ldif\"");
        
    }

    public function out(){
        echo $this->output;
    }


}
?>
