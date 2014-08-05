package roaring

import (
	. "github.com/smartystreets/goconvey/convey"
	"log"
	"testing"
)

func makeContainer(ss []uint16) container {
	c := newArrayContainer()
	for _, s := range ss {
		c.add(s)
	}
	return c
}

func checkContent(c container, s []uint16) bool {
	si := c.getShortIterator()
	ctr := 0
	fail := false
	for si.hasNext() {
		if ctr == len(s) {
			log.Println("HERE")
			fail = true
			break
		}
		i := si.next()
		if i != s[ctr] {

			log.Println("THERE", i, s[ctr])
			fail = true
			break
		}
		ctr++
	}
	if ctr != len(s) {
		log.Println("LAST")
		fail = true
	}
	if fail {
		log.Println("fail, found ")
		si = c.getShortIterator()
		z := 0
		for si.hasNext() {
			si.next()
			z++
		}
		log.Println(z, len(s))
	}

	return !fail
}

func TestRoaringContainer(t *testing.T) {
	Convey("NumberOfTrailingZeros", t, func() {
		x := int64(0)
		o := numberOfTrailingZeros(x)
		So(o, ShouldEqual, 64)
		x = 1 << 3
		o = numberOfTrailingZeros(x)
		So(o, ShouldEqual, 3)
	})
	Convey("ArrayShortIterator", t, func() {
		content := []uint16{1, 3, 5, 7, 9}
		c := makeContainer(content)
		si := c.getShortIterator()
		i := 0
		for si.hasNext() {
			si.next()
			i++
		}

		So(i, ShouldEqual, 5)
	})

	Convey("BinarySearch", t, func() {
		content := []uint16{1, 3, 5, 7, 9}
		res := binarySearch(content, 5)
		So(res, ShouldEqual, 2)
		res = binarySearch(content, 4)
		So(res, ShouldBeLessThan, 0)
	})
	Convey("bitmapcontainer", t, func() {
		content := []uint16{1, 3, 5, 7, 9}
		a := newArrayContainer()
		b := newBitmapContainer()
		for _, v := range content {
			a.add(v)
			b.add(v)
		}
		c := a.toBitmapContainer()

		So(a.getCardinality(), ShouldEqual, b.getCardinality())
		So(c.getCardinality(), ShouldEqual, b.getCardinality())

	})
	Convey("inottest0", t, func() {
		content := []uint16{9}
		c := makeContainer(content)
		c = c.inot(0, 10)
		si := c.getShortIterator()
		i := 0
		for si.hasNext() {
			si.next()
			i++
		}
		So(i, ShouldEqual, 10)
	})

	Convey("inotTest1", t, func() {
		// Array container, range is complete
		content := []uint16{1, 3, 5, 7, 9}
		//content := []uint16{1}
		edge := 1 << 13
		//		edge := 30
		c := makeContainer(content)
		c = c.inot(0, edge)
		size := edge - len(content)
		s := make([]uint16, size+1)
		pos := 0
		for i := uint16(0); i < uint16(edge+1); i++ {
			if binarySearch(content, i) < 0 {
				s[pos] = i
				pos++
			}
		}
		So(checkContent(c, s), ShouldEqual, true)
	})

	/*
	   @Test
	   public void inotTest10() {
	           System.out.println("inotTest10");
	           // Array container, inverting a range past any set bit
	           final uint16[] content = new uint16[3];
	           content[0] = 0;
	           content[1] = 2;
	           content[2] = 4;
	           final Container c = makeContainer(content);
	           final Container c1 = c.inot(65190, 65200);
	           assertTrue(c1 instanceof ArrayContainer);
	           assertEquals(14, c1.getCardinality());
	           assertTrue(checkContent(c1, new uint16[] { 0, 2, 4,
	                   (uint16) 65190, (uint16) 65191, (uint16) 65192,
	                   (uint16) 65193, (uint16) 65194, (uint16) 65195,
	                   (uint16) 65196, (uint16) 65197, (uint16) 65198,
	                   (uint16) 65199, (uint16) 65200 }));
	   }

	   @Test
	   public void inotTest2() {
	           // Array and then Bitmap container, range is complete
	           final uint16[] content = { 1, 3, 5, 7, 9 };
	           Container c = makeContainer(content);
	           c = c.inot(0, 65535);
	           c = c.inot(0, 65535);
	           assertTrue(checkContent(c, content));
	   }

	   @Test
	   public void inotTest3() {
	           // Bitmap to bitmap, full range

	           Container c = new ArrayContainer();
	           for (int i = 0; i < 65536; i += 2)
	                   c = c.add((uint16) i);

	           c = c.inot(0, 65535);
	           assertTrue(c.contains((uint16) 3) && !c.contains((uint16) 4));
	           assertEquals(32768, c.getCardinality());
	           c = c.inot(0, 65535);
	           for (int i = 0; i < 65536; i += 2)
	                   assertTrue(c.contains((uint16) i)
	                           && !c.contains((uint16) (i + 1)));
	   }

	   @Test
	   public void inotTest4() {
	           // Array container, range is partial, result stays array
	           final uint16[] content = { 1, 3, 5, 7, 9 };
	           Container c = makeContainer(content);
	           c = c.inot(4, 999);
	           assertTrue(c instanceof ArrayContainer);
	           assertEquals(999 - 4 + 1 - 3 + 2, c.getCardinality());
	           c = c.inot(4, 999); // back
	           assertTrue(checkContent(c, content));
	   }

	   @Test
	   public void inotTest5() {
	           System.out.println("inotTest5");
	           // Bitmap container, range is partial, result stays bitmap
	           final uint16[] content = new uint16[32768 - 5];
	           content[0] = 0;
	           content[1] = 2;
	           content[2] = 4;
	           content[3] = 6;
	           content[4] = 8;
	           for (int i = 10; i <= 32767; ++i)
	                   content[i - 10 + 5] = (uint16) i;
	           Container c = makeContainer(content);
	           c = c.inot(4, 999);
	           assertTrue(c instanceof BitmapContainer);
	           assertEquals(31773, c.getCardinality());
	           c = c.inot(4, 999); // back, as a bitmap
	           assertTrue(c instanceof BitmapContainer);
	           assertTrue(checkContent(c, content));

	   }

	   @Test
	   public void inotTest6() {
	           System.out.println("inotTest6");
	           // Bitmap container, range is partial and in one word, result
	           // stays bitmap
	           final uint16[] content = new uint16[32768 - 5];
	           content[0] = 0;
	           content[1] = 2;
	           content[2] = 4;
	           content[3] = 6;
	           content[4] = 8;
	           for (int i = 10; i <= 32767; ++i)
	                   content[i - 10 + 5] = (uint16) i;
	           Container c = makeContainer(content);
	           c = c.inot(4, 8);
	           assertTrue(c instanceof BitmapContainer);
	           assertEquals(32762, c.getCardinality());
	           c = c.inot(4, 8); // back, as a bitmap
	           assertTrue(c instanceof BitmapContainer);
	           assertTrue(checkContent(c, content));
	   }

	   @Test
	   public void inotTest7() {
	           System.out.println("inotTest7");
	           // Bitmap container, range is partial, result flips to array
	           final uint16[] content = new uint16[32768 - 5];
	           content[0] = 0;
	           content[1] = 2;
	           content[2] = 4;
	           content[3] = 6;
	           content[4] = 8;
	           for (int i = 10; i <= 32767; ++i)
	                   content[i - 10 + 5] = (uint16) i;
	           Container c = makeContainer(content);
	           c = c.inot(5, 31000);
	           if(c.getCardinality() <= ArrayContainer.DEFAULTMAXSIZE)
	                   assertTrue(c instanceof ArrayContainer);
	           else
	                   assertTrue(c instanceof BitmapContainer);
	           assertEquals(1773, c.getCardinality());
	           c = c.inot(5, 31000); // back, as a bitmap
	           if(c.getCardinality() <= ArrayContainer.DEFAULTMAXSIZE)
	                   assertTrue(c instanceof ArrayContainer);
	           else
	                   assertTrue(c instanceof BitmapContainer);
	           assertTrue(checkContent(c, content));
	   }

	   // case requiring contraction of ArrayContainer.
	   @Test
	   public void inotTest8() {
	           System.out.println("inotTest8");
	           // Array container
	           final uint16[] content = new uint16[21];
	           for (int i = 0; i < 18; ++i)
	                   content[i] = (uint16) i;
	           content[18] = 21;
	           content[19] = 22;
	           content[20] = 23;

	           Container c = makeContainer(content);
	           c = c.inot(5, 21);
	           assertTrue(c instanceof ArrayContainer);

	           assertEquals(10, c.getCardinality());
	           c = c.inot(5, 21); // back, as a bitmap
	           assertTrue(c instanceof ArrayContainer);
	           assertTrue(checkContent(c, content));
	   }

	   // mostly same tests, except for not. (check original unaffected)
	   @Test
	   public void notTest1() {
	           // Array container, range is complete
	           final uint16[] content = { 1, 3, 5, 7, 9 };
	           final Container c = makeContainer(content);
	           final Container c1 = c.not(0, 65535);
	           final uint16[] s = new uint16[65536 - content.length];
	           int pos = 0;
	           for (int i = 0; i < 65536; ++i)
	                   if (Arrays.binarySearch(content, (uint16) i) < 0)
	                           s[pos++] = (uint16) i;
	           assertTrue(checkContent(c1, s));
	           assertTrue(checkContent(c, content));
	   }

	   @Test
	   public void notTest10() {
	           System.out.println("notTest10");
	           // Array container, inverting a range past any set bit
	           // attempting to recreate a bug (but bug required extra space
	           // in the array with just the right junk in it.
	           final uint16[] content = new uint16[40];
	           for (int i = 244; i <= 283; ++i)
	                   content[i - 244] = (uint16) i;
	           final Container c = makeContainer(content);
	           final Container c1 = c.not(51413, 51470);
	           assertTrue(c1 instanceof ArrayContainer);
	           assertEquals(40 + 58, c1.getCardinality());
	           final uint16[] rightAns = new uint16[98];
	           for (int i = 244; i <= 283; ++i)
	                   rightAns[i - 244] = (uint16) i;
	           for (int i = 51413; i <= 51470; ++i)
	                   rightAns[i - 51413 + 40] = (uint16) i;

	           assertTrue(checkContent(c1, rightAns));
	   }

	   @Test
	   public void notTest11() {
	           System.out.println("notTest11");
	           // Array container, inverting a range before any set bit
	           // attempting to recreate a bug (but required extra space
	           // in the array with the right junk in it.
	           final uint16[] content = new uint16[40];
	           for (int i = 244; i <= 283; ++i)
	                   content[i - 244] = (uint16) i;
	           final Container c = makeContainer(content);
	           final Container c1 = c.not(1, 58);
	           assertTrue(c1 instanceof ArrayContainer);
	           assertEquals(40 + 58, c1.getCardinality());
	           final uint16[] rightAns = new uint16[98];
	           for (int i = 1; i <= 58; ++i)
	                   rightAns[i - 1] = (uint16) i;
	           for (int i = 244; i <= 283; ++i)
	                   rightAns[i - 244 + 58] = (uint16) i;

	           assertTrue(checkContent(c1, rightAns));
	   }

	   @Test
	   public void notTest2() {
	           // Array and then Bitmap container, range is complete
	           final uint16[] content = { 1, 3, 5, 7, 9 };
	           final Container c = makeContainer(content);
	           final Container c1 = c.not(0, 65535);
	           final Container c2 = c1.not(0, 65535);
	           assertTrue(checkContent(c2, content));
	   }

	   @Test
	   public void notTest3() {
	           // Bitmap to bitmap, full range

	           Container c = new ArrayContainer();
	           for (int i = 0; i < 65536; i += 2)
	                   c = c.add((uint16) i);

	           final Container c1 = c.not(0, 65535);
	           assertTrue(c1.contains((uint16) 3) && !c1.contains((uint16) 4));
	           assertEquals(32768, c1.getCardinality());
	           final Container c2 = c1.not(0, 65535);
	           for (int i = 0; i < 65536; i += 2)
	                   assertTrue(c2.contains((uint16) i)
	                           && !c2.contains((uint16) (i + 1)));
	   }

	   @Test
	   public void notTest4() {
	           System.out.println("notTest4");
	           // Array container, range is partial, result stays array
	           final uint16[] content = { 1, 3, 5, 7, 9 };
	           final Container c = makeContainer(content);
	           final Container c1 = c.not(4, 999);
	           assertTrue(c1 instanceof ArrayContainer);
	           assertEquals(999 - 4 + 1 - 3 + 2, c1.getCardinality());
	           final Container c2 = c1.not(4, 999); // back
	           assertTrue(checkContent(c2, content));
	   }

	   @Test
	   public void notTest5() {
	           System.out.println("notTest5");
	           // Bitmap container, range is partial, result stays bitmap
	           final uint16[] content = new uint16[32768 - 5];
	           content[0] = 0;
	           content[1] = 2;
	           content[2] = 4;
	           content[3] = 6;
	           content[4] = 8;
	           for (int i = 10; i <= 32767; ++i)
	                   content[i - 10 + 5] = (uint16) i;
	           final Container c = makeContainer(content);
	           final Container c1 = c.not(4, 999);
	           assertTrue(c1 instanceof BitmapContainer);
	           assertEquals(31773, c1.getCardinality());
	           final Container c2 = c1.not(4, 999); // back, as a bitmap
	           assertTrue(c2 instanceof BitmapContainer);
	           assertTrue(checkContent(c2, content));
	   }

	   @Test
	   public void notTest6() {
	           System.out.println("notTest6");
	           // Bitmap container, range is partial and in one word, result
	           // stays bitmap
	           final uint16[] content = new uint16[32768 - 5];
	           content[0] = 0;
	           content[1] = 2;
	           content[2] = 4;
	           content[3] = 6;
	           content[4] = 8;
	           for (int i = 10; i <= 32767; ++i)
	                   content[i - 10 + 5] = (uint16) i;
	           final Container c = makeContainer(content);
	           final Container c1 = c.not(4, 8);
	           assertTrue(c1 instanceof BitmapContainer);
	           assertEquals(32762, c1.getCardinality());
	           final Container c2 = c1.not(4, 8); // back, as a bitmap
	           assertTrue(c2 instanceof BitmapContainer);
	           assertTrue(checkContent(c2, content));
	   }

	   @Test
	   public void notTest7() {
	           System.out.println("notTest7");
	           // Bitmap container, range is partial, result flips to array
	           final uint16[] content = new uint16[32768 - 5];
	           content[0] = 0;
	           content[1] = 2;
	           content[2] = 4;
	           content[3] = 6;
	           content[4] = 8;
	           for (int i = 10; i <= 32767; ++i)
	                   content[i - 10 + 5] = (uint16) i;
	           final Container c = makeContainer(content);
	           final Container c1 = c.not(5, 31000);
	           if(c1.getCardinality() <= ArrayContainer.DEFAULTMAXSIZE)
	                   assertTrue(c1 instanceof ArrayContainer);
	           else
	                   assertTrue(c1 instanceof BitmapContainer);
	           assertEquals(1773, c1.getCardinality());
	           final Container c2 = c1.not(5, 31000); // back, as a bitmap
	           if(c2.getCardinality() <= ArrayContainer.DEFAULTMAXSIZE)
	                   assertTrue(c2 instanceof ArrayContainer);
	           else
	                   assertTrue(c2 instanceof BitmapContainer);
	           assertTrue(checkContent(c2, content));
	   }

	   @Test
	   public void notTest8() {
	           System.out.println("notTest8");
	           // Bitmap container, range is partial on the lower end
	           final uint16[] content = new uint16[32768 - 5];
	           content[0] = 0;
	           content[1] = 2;
	           content[2] = 4;
	           content[3] = 6;
	           content[4] = 8;
	           for (int i = 10; i <= 32767; ++i)
	                   content[i - 10 + 5] = (uint16) i;
	           final Container c = makeContainer(content);
	           final Container c1 = c.not(4, 65535);
	           assertTrue(c1 instanceof BitmapContainer);
	           assertEquals(32773, c1.getCardinality());
	           final Container c2 = c1.not(4, 65535); // back, as a bitmap
	           assertTrue(c2 instanceof BitmapContainer);
	           assertTrue(checkContent(c2, content));
	   }

	   @Test
	   public void notTest9() {
	           System.out.println("notTest9");
	           // Bitmap container, range is partial on the upper end, not
	           // single word
	           final uint16[] content = new uint16[32768 - 5];
	           content[0] = 0;
	           content[1] = 2;
	           content[2] = 4;
	           content[3] = 6;
	           content[4] = 8;
	           for (int i = 10; i <= 32767; ++i)
	                   content[i - 10 + 5] = (uint16) i;
	           final Container c = makeContainer(content);
	           final Container c1 = c.not(0, 65200);
	           assertTrue(c1 instanceof BitmapContainer);
	           assertEquals(32438, c1.getCardinality());
	           final Container c2 = c1.not(0, 65200); // back, as a bitmap
	           assertTrue(c2 instanceof BitmapContainer);
	           assertTrue(checkContent(c2, content));
	   }

	   @Test
	   public void rangeOfOnesTest1() {
	           final Container c = Container.rangeOfOnes(4, 10); // sparse
	           assertTrue(c instanceof ArrayContainer);
	           assertEquals(10 - 4 + 1, c.getCardinality());
	           assertTrue(checkContent(c, new uint16[] { 4, 5, 6, 7, 8, 9, 10 }));
	   }

	   @Test
	   public void rangeOfOnesTest2() {
	           final Container c = Container.rangeOfOnes(1000, 35000); // dense
	           assertTrue(c instanceof BitmapContainer);
	           assertEquals(35000 - 1000 + 1, c.getCardinality());
	   }

	   @Test
	   public void rangeOfOnesTest2A() {
	           final Container c = Container.rangeOfOnes(1000, 35000); // dense
	           final uint16 s[] = new uint16[35000 - 1000 + 1];
	           for (int i = 1000; i <= 35000; ++i)
	                   s[i - 1000] = (uint16) i;
	           assertTrue(checkContent(c, s));
	   }

	   @Test
	   public void rangeOfOnesTest3() {
	           // bdry cases
	           final Container c = Container.rangeOfOnes(1,
	                   ArrayContainer.DEFAULTMAXSIZE);
	           assertTrue(c instanceof ArrayContainer);
	   }

	   @Test
	   public void rangeOfOnesTest4() {
	           final Container c = Container.rangeOfOnes(1,
	                   ArrayContainer.DEFAULTMAXSIZE + 1);
	           assertTrue(c instanceof BitmapContainer);
	   }

	   public static boolean checkContent(Container c, uint16[] s) {
	           ShortIterator si = c.getShortIterator();
	           int ctr = 0;
	           boolean fail = false;
	           while (si.hasNext()) {
	                   if (ctr == s.length) {
	                           fail = true;
	                           break;
	                   }
	                   if (si.next() != s[ctr]) {
	                           fail = true;
	                           break;
	                   }
	                   ++ctr;
	           }
	           if (ctr != s.length) {
	                   fail = true;
	           }
	           if (fail) {
	                   System.out.print("fail, found ");
	                   si = c.getShortIterator();
	                   while (si.hasNext())
	                           System.out.print(" " + si.next());
	                   System.out.print("\n expected ");
	                   for (final uint16 s1 : s)
	                           System.out.print(" " + s1);
	                   System.out.println();
	           }
	           return !fail;
	   }

	*/

}
