add(a,b,c) {
    auto n;
    n=a+b; n=n+c;
    return n;
}

endl() {
    putchar('\n');
}

main() {
    auto x; auto y; auto z;
    x=1; y=2; z=3;
    auto i; i=add(x,y,z);
    putnumb(i); endl();
}

