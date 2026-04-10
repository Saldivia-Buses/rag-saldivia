<?php
/**
 * input Text Box
 * @author Luis M. Melgratti
 * Created on 19/01/2008
 *
 */
class geoMap extends Html_input {


    function __construct($valor = '', $type ='') {

        $this->value 	= $valor;
        $this->type 	= $type;
        
    }

    function show() {

        $style = '';
        $size		= $this->size;

        $div = '';
        $maxsize = '';
        $valor = $this->value;

        $this->addParameter('value', $valor);

        if ($this->valorCampo != '')
            $this->addParameter('value', $this->valorCampo);

        if(isset( $this->maxsize) ) $maxsize = $this->maxsize;
        $this->Parameters['id'] = 'GEOMAP'.$this->Parameters['id'] ;
        $eventos 	= $this->getEventsString();
        $atributos 	= $this->getParametersString();
        $this->id = $this->Parameters['id'];
        $strmaxsize = ($maxsize != '')?  ' maxlength="'.$maxsize.'" ' :'';
//        $input = '<input '.$atributos.' '.$eventos .' '.$this->tabindex.' size="'.$size.'"  '.$strmaxsize.'/>';

        $style ='style="height:'.$size.'px;width:'.$size.'px"' ;

        $map = '<div '.$atributos.' '.$style.' >Loading Map</div>';

        $map .= $this->jsInit();

        return $map;
    }

    private function jsInit() {
        $iniLat = '-32.945616471035706';
        $iniLon = '-60.63649535179138';

        $iniLat = '-33.03062';
        $iniLon = '-60.61451';
        
        $js = "<script type=\"text/javascript\">

        var latlng = new google.maps.LatLng($iniLat, $iniLon);
         var my_map = $('#".$this->id."');
         my_map.css({width:'500px', height:'500px'});
        var myOptions = {
          zoom: 15,
          center: latlng,
          mapTypeId: google.maps.MapTypeId.HYBRID
        };
        var map = new google.maps.Map(my_map.get(0), myOptions);
        Histrix.mymap = map;

        
        </script>";
        loger($js);
        return $js;
    }

    private function jsInit_v2(){
        $iniLat = '-32.945616471035706';
        $iniLon = '-60.63649535179138';
        $js = "<script type= \"text/javascript\">
        if (GBrowserIsCompatible()) {
            Histrix.map['".$this->id."'] = new GMap2($('#".$this->id."')[0]);
            Histrix.map['".$this->id."'].setUIToDefault();
            var point = new GLatLng(\"".$iniLat."\",\"".$iniLon."\");
         //   var marker = new GMarker(point, {title: \"Salida\"});
          //  Histrix.map['".$this->id."'].addOverlay(marker);
            Histrix.map['".$this->id."'].setCenter(point, 17);

            // NEW POLYGON
            Histrix.geometryControls = new GeometryControls();

            ";
        if ($this->type == 'geoPoint'){
            $js .= "Histrix.markerControl['".$this->id."'] = new MarkerControl();";
            $js .= "Histrix.geometryControls.addControl(Histrix.markerControl['".$this->id."'] );";

        }
        if ($this->type == 'geoPoly'){
            $js .= "Histrix.polygonControl['".$this->id."'] = new PolygonControl();";
            $js .= "Histrix.geometryControls.addControl(Histrix.polygonControl['".$this->id."'] );";
        }
        

        $js .= "Histrix.map['".$this->id."'].addControl(Histrix.geometryControls , new GControlPosition(G_ANCHOR_BOTTOM_RIGHT));
            
            
        }
        else {
            Histrix.alerta('Navegador Incompatible para mostrar Mapas');
        }
        </script>";
        loger($js);
        return $js;
    }

}
?>
