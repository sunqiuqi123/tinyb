add(a,b) {
    return a+b;
}

main() {
    putchar('H'); putchar('e');
    auto i; i=2;
    while (i>0) {
        putchar('l');
	i=i-1;
    }
    putchar('o'); putchar('\n');
    auto x;
    x=getchar();
    auto y; y=add(x,1);
    putnumb(y);
}

