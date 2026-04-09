<?php
/* 
 * FieldType Class
 * 
 */

/**
 * Define FieldType representation
 *
 * @author luis
 */
class FieldType_editor extends FieldType_varchar{

    const ALIGN   = 'left'; // Default Alignment
    const DIR     = 'ltr';  // Text direction
    const INPUT   = 'editor';  // input type


    public function __construct(&$field=null){
        $this->field = $field;
    }

	public static function renderInput( $valor, $field, $arrayAtributos, $uiClass, $opciones=''){

        if ($arrayAtributos['disabled'] == 'disabled' || $arrayAtributos['readonly'] == 'readonly') {
            $salida = $valor;

        } else {
            $inputBox = new Html_textArea();
            $inputBox->value      = $valor;

            $inputBox->Parameters = $arrayAtributos;

            $inputBox->addParameter('maxlength', $field->maxlength);

            $inputBox->addParameter('internal_class', 'simpleditor');

            $inputBox->addStyle('width', '90%');

            // add custom Javascript Events
            if ($field->jsfunction)
                foreach ($field->jsfunction as $jsevent => $jsfunctions) {
                    foreach($jsfunctions as $nfunc => $jsfunction)
                        $inputBox->addEvent($jsevent, $jsfunction, true); // append function
                }

            $salida = "<div>" . $inputBox->show() . "</div>";
        }

        return $salida;

    }
}
?>
