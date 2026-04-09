<?php
include ("./sessionCheck.php");

//Search the ISBN database for the book.
    $isbn = $_POST['ISBN'];


    //$isbn = '9789505640614';
    $url = "http://www.isbndb.com/api/books.xml?access_key=BLBK666P&results=texts&index1=isbn&value1=".$isbn;

    $XMLabm = simplexml_load_file($url);
	$BookData 	= $XMLabm->BookList->BookData;
	$title 		= $BookData->Title;
	$titlelong 	= $BookData->TitleLong;
	$author 	= $BookData->AuthorsText;
	$publisher 	= $BookData->PublisherText;
	$summary 	= utf8_decode($BookData->Summary);
	$notes 		= $BookData->Notes;

	$isbn10		= $BookData['isbn'];


?>
<fieldset>
  <table>
      <tr>
      	<td>Isbn:</td><td><?php echo $isbn?></td>
    </tr>

    <tr>
      <td>T&iacute;tulo:</td><td><?php echo $title?></td>
    </tr>
	<tr>
      <td>T&iacute;tulo Largo:</td><td><?php echo $titlelong ?></td>
    </tr>
    <tr>
      <td>Autor:</td><td><?php echo $author; ?></td>
    </tr>
    <tr>
      <td>Editorial:</td><td><?php echo $publisher?></td>
    </tr>
    <tr>
      <td>Sumario:</td><td><?php echo htmlentities($summary);?></td>
    </tr>
    <tr>
      <td>Notas:</td><td><?php echo $notes;?></td>
    </tr>

  </table>
</fieldset>
