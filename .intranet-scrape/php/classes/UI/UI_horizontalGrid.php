<?php
/* 
 * 2009-09-09
 * basic table for information display
 */


class UI_horizontalGrid extends UI_consulta {

/**
 * User Interfase constructor
 *
 */
    public function __construct($Datacontainer) {
        parent::__construct($Datacontainer);

        $this->muestraCant = true;


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

        if ( $this->Datos->tipoAbm == 'ing' || $this->Datos->tipoAbm == 'grid')
            $abming = true;

        $llenoTemporal = (isset($this->Datos->llenoTemporal)) ? $this->Datos->llenoTemporal:'';
        $preload       = (isset($this->Datos->preloadData))?$this->Datos->preloadData:'';

        $this->TIEMPO_CONSULTA= processing_time();

        if ($llenoTemporal != "false" && $segundaVez == '' && $opt !='noselect') {
            if ( $this->nosel == 'true') {
            // 'no hago select';

            }else {
                if ($preload != "false") {
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

        /* Show an Horizontal Grid */
        return $this->showHorizontalGrid($opt, $parentObject);

    }


    public function showHorizontalGrid($opt, &$parentObject=null ) {
        $salida = '';
        $form = 'Form'.$this->Datos->idxml;

        if ($this->Datos->xmlpadre != '')
            $form = 'Form'.$this->Datos->xmlpadre;

        $originalFieldName= $parentObject->NombreCampo;
        //$this->Datos->calculointerno();
        $Tablatemp = $this->Datos->TablaTemporal->datos();


        // Generate Column labels
        $i = 0;
        if ($Tablatemp)
            foreach($Tablatemp as $order => $row) {
                $i++;
                foreach($row as $fieldName => $value) {

                    $field = $this->Datos->getCampo($fieldName);

                    if (isset($field->columnLabel) && $field->columnLabel =='true')
                        $header .= '<th>'.htmlentities(ucfirst($value), ENT_QUOTES, 'UTF-8').'</th>';

                    if (isset($field->Oculto))continue;

                    $prefix = $originalFieldName .'['.$fieldName.']'. '['.$i.']';
                    $atrib['arrayorder'] = $i;
                    $atrib['basename'] = $originalFieldName .'['.$fieldName.']';
                    $rowLabel[$fieldName] = $field->Etiqueta;

                    if (isset($field->valAttribute) && $field->valAttribute != '') {
                        foreach($field->valAttribute as $attribID => $attrib) {

                            $valAtt = $row[$attrib];

                            if ($attribID == 'oculto') {
                                if ($valAtt === 'true' || $valAtt === true)
                                    $field->setOculto($valAtt);
                            }

                            $field->{$attribID} = (string) $valAtt;
                            $field->Parametro[$attribID]= (string) $valAtt;

                        }
                    }

                    if ($opt == 'micro') {

                        if (isset($this->Datos->innerTablaData) && $this->Datos->innerTablaData != '') {
                            $cellValue =  $this->Datos->innerTablaData[0][$fieldName][$i];
                        }
                        else {
                            $cellValue = $value;
                        }

                        if (isset($field->paring) && $field->paring != '') {
                            $parametros .= $this->generateLinkParameters($field, $row);
                        }

                        $td = $field->renderCell($this , $prefix, $cellValue, $order, 0 , 0, $parametros);
                        if (isset($inputs[$fieldName]))
                             $inputs[$fieldName] .= $td;
                        else $inputs[$fieldName]  = $td;
                    }
                    else {
                        $inputs[$fieldName] .= '<td>'.$field->renderInput($this, $form, $prefix , null, '',  '', $atrib).'</td>';
                    }
                }
            }
        //$this->Datos->innerTablaData = $innerTable;
        unset ($this->Datos->innerTablaData);

        if ($i > 0){
            $salida = '<table class="autofields" style="width:100%">';
            // headers
            if (isset($header)){             
                $salida .= '<thead>';
                $salida .= '<tr><th></th>'.$header.'</tr>';
                $salida .= '</thead>';
            }
            // Generate Input Fields
            $salida .= '<tbody>';
            if ($inputs)
                foreach($inputs as $nom => $valin) {
                    
                $rowLbl = ($rowLabel[$nom] != '')?'<th>'.$rowLabel[$nom].'</th>':'';
                $salida .= '<tr id="'.$nom.'">'.$rowLbl.$inputs[$nom].'</tr>';
                }
            $salida .= '</tbody>';
            $salida .= '</table>';
        }
        return $salida;
    }


}


?>
