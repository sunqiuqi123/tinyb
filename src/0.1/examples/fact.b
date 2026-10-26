fact(i) {
    auto m,n; m=0; n=0;
    while(n<i) {
        n=n+1;
        m=m+n;
    }
    return m;
}

main() {
    auto x; x=5;
    auto y; y=fact(x);
    putnumb(y); putchar('\n');
}