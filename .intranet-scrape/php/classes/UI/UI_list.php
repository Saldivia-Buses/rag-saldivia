<?php
/* 
 * 2009-09-09
 * help popup class 
 */

class UI_list extends UI_consulta {

/**
 * User Interfase constructor
 *
 */
    public function __construct(&$Datacontainer) {
        parent::__construct($Datacontainer);
        $this->tag = 'div';
        $this->hasForm = true;
        $this->rowTag = 'li';
        $this->rowClass = 'list gradientDarkLight ui_roundCorners';
    }

    public function showTablaInt($opt = '', $idTabla = '', $segundaVez = '', $nocant='', $div=false, $form=null, $pdf=null, &$parentObject=null) {

        $defaultForm = 'Form'.$this->Datos->idxml;

        // nombre del form
        if ($form == null ) {
            $form = $defaultForm;
        }

        $form = str_replace('.', '_', $form);

        // Si es un subForm interno estos valores NO coinciden y no escribo el tag form
        if ($form == $defaultForm && $this->Datos->isInner != 'true') {
            $formini = '<form id="'.$form.'" name="'.$form.'" onsubmit="return false;" action="">';
            $formfin = '</form>';
        }

        $salida = '';

        $llenoTemporal = $this->Datos->llenoTemporal;

        $this->TIEMPO_CONSULTA= processing_time();

        if ($llenoTemporal != "false" && $segundaVez == '' && $opt !='noselect') {
            if ( $this->nosel == 'true') {
            // 'no hago select';

            }else {
                if ($this->Datos->preloadData != "false") {
                    $this->Datos->Select();
                }
                $this->Datos->preloadData = "true";

                if ($this->Datos->resultSet)
                    $this->cantCampos = _num_fields($this->Datos->resultSet);
                else $campos = $this->cantCampos();

                // Cargo tabla temporal con el resultado del select ODBC
                // Tarda un poco mas, SI, pero despues lo trato mas facil en la temporal :D
                // Y puedo Paginar sin tener en cuenta restricciones en el motor SQL
                // YA SE que es mas lento, pero bueno, velocidad x interoperabilidad
                // Que se le va a hacer...
                
                $this->Datos->CargoTablaTemporal();

            }
        }


        $contenido = $this->showDatos($idTabla, $opt);

        if ($this->Datos->sortable == 'true')
                $sortclass= ' sortablelist ';
 //          if (isset($this->Datos->swap) && $this->Datos->swap == 'true') $tableClass .= 'dnd';
//                 $salida .= '<table class="sortable resizable '.$tableClass.'"
//                        id="TablaInterna'.$idTabla.'"  width="100%" cellspacing="0" '.$styleTable.' '.$tableProp.'>';

        $salida = '<div id="'.$this->Datos->idxml.'"><ul class="ullist '.$sortclass.$tableClass.'" id="TablaInterna'.$idTabla.'" xml="'.$this->Datos->xml.'" instance="'.$this->Datos->getInstance().'">'.$contenido.'</ul></div>';

        return $salida;
    }




}

?>