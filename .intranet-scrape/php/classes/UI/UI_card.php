<?php
/* 
 * 2009-09-09
 * help popup class 
 */

class UI_card extends UI_list {

/**
 * User Interfase constructor
 *
 */
    public function __construct(&$Datacontainer) {
        parent::__construct($Datacontainer);
        $this->tag = 'div';
        $this->hasForm = true;
        $this->rowTag = 'div';
        $this->rowClass = 'card gradientDarkLight ui_roundCorners';
    }

    // pdf printing of data 
    public function pdf($pdf , $fontsize = '', $opImpresion ='',$anchoTabla='', $posx=''){
//        $pdf->SetY(30);
        $pdf->SetFontSize($fontsize);        
        $Tablatemp = $this->Datos->TablaTemporal->datos();   

        $width  = 30;
//        $height = 10;

        
        if (isset($this->Datos->grid)){ 
            $gridValues = explode(',' , $this->Datos->grid);
            //$gridStyle = ' style="width:'.$gridValues[0].';height:'.$gridValues[1].';" ';
            
            // calculates widths for cards
            if (strpos($gridValues[0], '%')){
                $porc = str_replace('%','',$gridValues[0]);
                $width = $pdf->anchoPagina / 100 * $porc + 4;
                $height  = $width;
            }
            
            
        }        
        
        
        //reset X position
        $leftMargin = 5;
        $posX = $leftMargin;
        
        $margin = 1;
        $padding = 2;
        $height = 0;
        
        // get Rows
        if ($Tablatemp != ''){
            foreach ($Tablatemp as $orden => $row) {
                $rownum++;        
                
                $posY = $pdf->GetY();

                // check if page width ends and skip a row
                if (($posX + $width) > $pdf->anchoPagina){
                    $posX = $leftMargin;
                    $posY += $height + $margin;
                    $height = 0;
                    $pdf->SetXY($leftMargin, $posY);
                    
                }

                // check if page ends and skip a page
                if ($pdf->CheckPageBreak($height)){
                    $posY = 30;
                }
                
                
                $innerX = $posX ;
                $innerY = $posY + $padding;

                // Write the content

                foreach($row as $fieldName => $value){
                    $field = $this->Datos->getCampo($fieldName);
                    if (!is_object($field))
                        continue;        
                    // Remove Non printables
                    if (isset($field->Oculto) && ($field->Oculto || $field->Oculto == 'true'))
                        continue;
                    if (isset($field->noshow) && $field->noshow == 'true')
                        continue;
                    if (isset($field->colstyle) && strpos('_' . $field->colstyle, 'display:none'))
                        continue;
                    if (isset($field->noEmpty) && $field->noEmpty == 'true' && !isset($this->Datos->hasValue[$field->NombreCampo])) {
                        continue;
                    }                    
                    
                    // setValue


                    if (isset($field->contExterno)) {
                        $field->Parametro['pdfwidth'] = $width;
                        
                        $field->PDFwidth = $width ;
                
                        $field->refreshInnerDataContainer($this->Datos, $row);

                    }
                    // set object position
                    $field->posX = $innerX ;
                    $field->posx = $innerX ;
                    $field->posy = $pdf->maxY;
                    $pdf->lastY  = $innerY;
                    $lastLabelY = $pdf->labelXY($field, true, $this->Datos, $value);      
                    
                 //   $pdf->loger($lastLabelY, 30,$lastLabelY );
		    $pdf->SetY($lastLabelY);                    
                    
                }

                // Draw the card
                $pdf->SetDrawColor(192, 192, 192);
                
                $height= max($height, $lastLabelY - $posY);
                $pdf->RoundedRect($posX, $posY, $width , $height, 1);

                // reset Position
                $pdf->SetXY($posX, $posY);
                $pdf->maxY = $innerY;        
                // advance to next Card
                $posX += $width + $margin;
                    
            }
        }
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
