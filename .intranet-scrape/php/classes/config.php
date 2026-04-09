<?php
/**
 * Leo la configuracion del xml
 */

class config {
    var $xml;
    var $path;
    var $nombre;
    var $base_default;
    var $base_actual;
    var $xmlPath;
    // Array con las bases disponibles
    var $bases;

    var $nom_empresa;

    function config($xml='', $path=DBDIR, $db='') {
        $this->xml = $xml;
        if ($path != '')
            $this->path = $path;
        else $this->path = 'database/';

        // for use with custom XML config files
        $customxml =  $db.'_cfg.xml';
        if (is_file(CFGFILEDIR.$customxml)){
            $this->path = CFGFILEDIR;
            $this->xml = $customxml;
        }


        if ($this->xml != '')
            $this->leo_xml($db);

    }

    function leo_xml($db='') {
    
        $xmlfile = $this->path.$this->xml;
        if(!file_exists($xmlfile)) die(ERR001);
        
        $XML = @simplexml_load_file($xmlfile);

        if (!$XML) die(ERR002);
        foreach ($XML->empresa as $empresa) {
            $this->nom_empresa 	= (string) $empresa->nombre;
            $this->imapServer 	= (string) $empresa->imapServer;
            $this->emailProgramm= (string) $empresa->emailProgramm;
            $this->supportUrl= (string) $empresa->supportUrl;
	    
	    
            $this->lang	 	= (string) $empresa->lang;
            $this->direccion 	= (string) $empresa->direccion;
            $this->cuit 	= (string) $empresa->cuit;
            $this->telefonos   	= (string) $empresa->telefonos;

            $this->logo_pdf_1  = (string) $empresa->logo_pdf_1;
            $this->logo_pdf_2  = (string) $empresa->logo_pdf_2;
			
            $this->css    	= (string) $empresa->css;
            $this->logo_ini    	= (string) $empresa->logo_ini;
            $this->img_fondo   	= (string) $empresa->img_fondo;
        }

        foreach ($XML->conexiones as $conexiones) {

            $this->base_default = (string) $conexiones['default'];
            $this->base_actual = $this->base_default;

            foreach ($conexiones as $base) {
                $id = (string) $base['id'];
                $DB = new base($id);

//                if ($db != '' && $db != $id) continue;

                $DB->tipo 	  = (string) $base['tipo'];
                $DB->xmlPath 	  = (string) $base['xmlPath'];

                $DB->lang	  = (string) $base->lang;
                $DB->descripcion  = (string) $base->descripcion;
                $DB->dsn	  = (string) $base->dsn;
                $DB->base	  = (string) $base->base;
                $DB->driver	  = (string) $base->driver;
                $DB->user 	  = (string) $base->user;
                $DB->password	  = (string) $base->password;
                $DB->host 	  = (string) $base->host;
                $DB->port     = (string) $base->port;                
                $DB->gmapKey	  = (string) $base->gmapKey;
                $this->bases[$id] = $DB;


                // defaults values
                $DB->nom_empresa  = $this->nombre;
                $DB->imapServer   = $this->imapServer;
                if (isset($this->emailProgram))
                    $DB->emailProgram = $this->emailProgram;
                
                $DB->lang	  = $this->lang;
                $DB->direccion    = $this->direccion;
                $DB->cuit         = $this->cuit;
                $DB->telefonos    = $this->telefonos;

                $DB->logo_pdf_1   = $this->logo_pdf_1;
                $DB->logo_pdf_2   = $this->logo_pdf_2;

                $DB->css    	  = $this->css;
                $DB->logo_ini     = $this->logo_ini;
                $DB->img_fondo    = $this->img_fondo;


                // Read Parameters
                foreach ($base->empresa as $empresa)
                {
                    foreach($empresa as $nompar => $valpar) 
                    {
                        $value = (string) $valpar;

                        $DB->{$nompar} = $value;

		                foreach ($valpar->attributes() as $attrname => $attr_value) {
		                    $DB->{$nompar.'_attr'}[$attrname] = (string) $attr_value;
		                }
                        // as safe as it can be
                        $value = addslashes(htmlentities($value));

                        $DB->properties[$nompar] = $value;
                    }                   
                }
                
            }
        }
    }
}
?>
