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
class FieldType_dir extends FieldType
{
    const ALIGN = 'left'; // Default Alignment
    const DIR   = 'ltr';  // Text direction

    public function __construct(&$field=null)
    {
        $this->field = $field;
    }

    public static function customValue($value, $field, $parameters='')
    {
        if ($value != '') {
            $this_field = $field;

            $orden = $parameters['order'];
            $basePath = '../database/'.$_SESSION['datapath'].'xml/';
    
	    $basedir=$value;

	    $slashdir = '';

	    if ($this_field->showButtom =="true" ){
		
	    } else {
    		$fileManager = new FileManager ($slashdir, $basePath, 'r');
    	        $fileManager->basedir    = $basedir;
                $images = $fileManager->getDirContents( $basePath, $basedir,  1 );


	        if (is_array($images)) {
    	            foreach ($images as  $imageurl) {
        	        // $url =  $basePath. $basedir.$imageurl;
            	        $file = new Archivo($imageurl, $basedir, '', $this_field);
                	$value2 .= $file->showInline();
                    }

	            $value = $value2;
    			} else $value = '';
	    }
	    return $value;
    	}
    }
}
