<?php

/**
 * Notification Class
 *
 * @author Luis M. Melgratti
 * @creation 2009-09-19
 */
class Postit extends Notification {

    public function __construct($id, $title='', $text='', $fade= 0, $class= 'Postit') {
        $this->id = $id;
        $this->title = $title;
        $this->contentString = $text;
//        $this->class = $notification;
        $this->fade = $fade;
        $this->readCondition = 2;
    }

    protected function processResults($row) {
        $uid = $row['idMensaje'];

        $date = Types::formatDate($row['fecha'], 'd/m/y');
        $user = $row['Nombre'] . ' ' . $row['apellido'];

        $header = '<table width="100%"><tr><td class="name"><img   src="../img/question.gif" />  ' . $user . '</td><td class="date" >' . $date . ' ' . substr($row['hora'], 0, 5) . '</td></tr></table>';
        $message = '<span>' . $row['mensaje'] . '</span>';
        //<br><span class="key">' . $key . '</span>';

        $message = nl2br(addslashes($message));
        $message = trim(preg_replace('/\s+/', ' ', $message));

        $notification = new Postit($uid, $header, $message, 0);
        $notification->onClick("$.post('sysmon.php', {'deletePostit':" . $row['idMensaje'] . "});");

        echo $notification->output();
    }

    public function createTag() {
        $this->code[] = "Histrix.Postit('$this->id', {title:'$this->title', text:'$this->contentString', notifClass:'$this->class', click: function(){ $this->click }, fade:$this->fade });";
    }

}

?>
