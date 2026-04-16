<?php
class RegistroEx {

    var $id;
    var $xml;
    var $campo; // guarda los valores a insertar/Actualizar

    var $par;   // guarda las relaciones
    function RegistroEx($id, $xml) {
        $this->id = $id;
        $this->xml = $xml;
    }
    // Recorre los campos buscando por este contenedor y lo actualiza y refresca
    public function Actualizar($ContenedorBase, $del) {

        //Busco el Contenedor dentro de cada Campo
        // Primero en la cabecera
        if (isset ($ContenedorBase->CabeceraMov))
            foreach ($ContenedorBase->CabeceraMov as $NCabecera => $ContCab) {

                $done = false;
                foreach ($ContCab->tablas[$ContCab->TablaBase]->campos as $numCamCab => $CampoCab) {
                    if ($CampoCab->contExterno != '')
                        if ($CampoCab->contExterno->xml == $this->xml) {
                            if ($this->campo!='')
                                $this->GraboContExterno($CampoCab->contExterno, $del);
                        }

                }

            }

        //Luego en el propio?
        // Actualizo su tabla temporal

        // refresco El contenedor Encontrado
    }

    function GraboContExterno($ContExterno, $del) {
        $this->campo['_AUTO']=true;
        $xmlref = $ContExterno->xml;
        $contReferente = new ContDatos("");
        $contReferente = Histrix_XmlReader::unserializeContainer($ContExterno);

        if ($del)
            $contReferente->TablaTemporal->deleteAutos();
        $contReferente->TablaTemporal->insert($this->campo);
        $contReferente->calculointerno();

        Histrix_XmlReader::serializeContainer($contReferente);

        $xml= $this->xmlpadre;
        $optionsArray['instance'] = '"'.$contReferente->getInstance().'"';
        $postOptions = Html::javascriptObject($optionsArray,'"');
//        $postOptions = htmlspecialchars($postOptions);
        
        // refresco referente
        $reload .= 'grabaABM(\'Form'.$xmlref.'\', \'update\', \''.$xmlref.'\' , \''.$xmlref.'\', null, null, '.$postOptions.' ); ';
//        echo $reload;
        echo '<script type="text/javascript">';
        echo $reload;
        echo '</script>';

    }

}
?>